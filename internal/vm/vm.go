/*
 *
 * Chippy - internal/vm/vm.go
 *
 */

package vm

import (
	"chip-go/internal/bytecode"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"chip-go/internal/vm/collections"
	"fmt"
	"math"
)

// callFrame is one call that is currently running. Its local variables are kept
// here rather than on the value stack, so one captured by a function defined
// inside it keeps the same address for as long as that function holds it.
type callFrame struct {
	closure *Function
	chunk   *bytecode.Chunk
	ip      int
	locals  []values.Value

	stackBase int
	forDepth  int

	localsMark int
}

const noLocalsMark = -1

// forState is one for-loop that is currently running. done means the next step
// would take the counter past the largest whole number that fits, so the loop
// must stop.
type forState struct {
	isInt bool
	done  bool

	intCounter, intEnd, intStep       int64
	floatCounter, floatEnd, floatStep float64
}

func (s *forState) inRange() bool {
	if s.isInt {
		return !s.done && ((s.intStep >= 0 && s.intCounter <= s.intEnd) || (s.intStep < 0 && s.intCounter >= s.intEnd))
	}

	return (s.floatStep >= 0 && s.floatCounter <= s.floatEnd) || (s.floatStep < 0 && s.floatCounter >= s.floatEnd)
}

func (s *forState) counter() (values.Value, error) {
	if s.isInt {
		return values.NewNumber(s.intCounter), nil
	}

	return values.NewNumberFromFloat(s.floatCounter)
}

func (s *forState) advance() {
	if !s.isInt {
		s.floatCounter += s.floatStep

		return
	}

	if (s.intStep > 0 && s.intCounter > math.MaxInt64-s.intStep) || (s.intStep < 0 && s.intCounter < math.MinInt64-s.intStep) {
		s.done = true

		return
	}

	s.intCounter += s.intStep
}

type VM struct {
	stack    []values.Value
	frames   []callFrame
	forStack []forState
	ctx      values.Ctx
	globals  *context.Globals[values.Value]

	locals    []values.Value
	localsTop int
}

func NewVM(chunk *bytecode.Chunk, ctx values.Ctx) *VM {
	if ctx.Globals == nil {
		errors.ModenaPanic("context has no globals")
	}

	vm := &VM{ctx: ctx, globals: ctx.Globals}
	vm.stack = make([]values.Value, 0, constants.SIZE_STACK_INITIAL)
	vm.frames = append(vm.frames, callFrame{chunk: chunk, locals: make([]values.Value, chunk.NumSlots)})

	return vm
}

func (vm *VM) Run() (values.Value, error) {
	frame := &vm.frames[len(vm.frames)-1]
	chunk := frame.chunk
	code := chunk.Code
	ip := 0

	for ip < len(code) {
		op := bytecode.Op(code[ip])
		opStart := ip
		ip++

		var err error

		switch op {
		case bytecode.OpConstant:
			index := bytecode.ReadU32(code, ip)
			ip += 4

			vm.push(chunk.Constants[index])

		case bytecode.OpDiscard:
			vm.discard()

		case bytecode.OpGetGlobal:
			nameIndex := int(bytecode.ReadU32(code, ip))
			ip += 4

			value := vm.globals.SlotValue(chunk.GlobalSlot(nameIndex))

			if value.IsUnset() {
				span := chunk.SpanAt(opStart)

				return values.Value{}, errors.NewRTError(span.Start, span.End, fmt.Sprintf("'%s' is not defined", chunk.Names[nameIndex]))
			}

			vm.push(value)

		case bytecode.OpDefineGlobal:
			nameIndex := int(bytecode.ReadU32(code, ip))
			ip += 4

			vm.globals.SetSlot(chunk.GlobalSlot(nameIndex), vm.peek())

		case bytecode.OpSetGlobal:
			nameIndex := int(bytecode.ReadU32(code, ip))
			ip += 4

			slot := chunk.GlobalSlot(nameIndex)

			if vm.globals.SlotValue(slot).IsUnset() {
				span := chunk.SpanAt(opStart)

				return values.Value{}, errors.NewRTError(span.Start, span.End, fmt.Sprintf("'%s' is not defined", chunk.Names[nameIndex]))
			}

			vm.globals.SetSlot(slot, vm.peek())

		case bytecode.OpJump:
			ip = int(bytecode.ReadU32(code, ip))

		// Reads the top value without taking it, so 'and' and 'or' can
		// leave their result behind for the compiler to discard.
		case bytecode.OpJumpIfFalse:
			target := bytecode.ReadU32(code, ip)
			ip += 4

			if !vm.peek().IsTrue() {
				ip = int(target)
			}

		case bytecode.OpGetLocal:
			slot := int(bytecode.ReadU32(code, ip))
			ip += 4

			value := frame.locals[slot]

			if value.IsUnset() {
				span := chunk.SpanAt(opStart)

				return values.Value{}, errors.NewRTError(span.Start, span.End, fmt.Sprintf("'%s' is not defined", chunk.LocalNames[slot]))
			}

			vm.push(value)

		case bytecode.OpSetLocal:
			slot := int(bytecode.ReadU32(code, ip))
			ip += 4

			frame.locals[slot] = vm.peek()

		case bytecode.OpGetUpvalue:
			index := int(bytecode.ReadU32(code, ip))
			ip += 4

			value := *frame.closure.upvalues[index]

			if value.IsUnset() {
				span := chunk.SpanAt(opStart)

				return values.Value{}, errors.NewRTError(span.Start, span.End, fmt.Sprintf("'%s' is not defined", frame.closure.template.Upvalues[index].Name))
			}

			vm.push(value)

		case bytecode.OpSetUpvalue:
			index := int(bytecode.ReadU32(code, ip))
			ip += 4

			*frame.closure.upvalues[index] = vm.peek()

		case bytecode.OpClosure:
			template := chunk.Functions[bytecode.ReadU32(code, ip)]
			ip += 4

			fn := &Function{template: template, globals: vm.ctx}
			fn.upvalues = make([]*values.Value, len(template.Upvalues))

			// A variable captured from the surrounding call points
			// straight at that call's own copy. One captured from
			// further out is shared with the surrounding function,
			// which already holds it.
			for i, desc := range template.Upvalues {
				if desc.FromLocal {
					fn.upvalues[i] = &frame.locals[desc.Index]
				} else {
					fn.upvalues[i] = frame.closure.upvalues[desc.Index]
				}
			}

			vm.push(values.NewFunctionValue(fn))

		case bytecode.OpCall:
			argCount := int(bytecode.ReadU32(code, ip))
			ip += 4

			calleeIdx := len(vm.stack) - 1 - argCount
			fn, isChippyFunc := vm.stack[calleeIdx].FunctionObject().(*Function)

			if !isChippyFunc {
				if err := vm.callBuiltin(calleeIdx, argCount, chunk, opStart); err != nil {
					return values.Value{}, err
				}

				break
			}

			if err := vm.checkCallDepth(chunk, opStart); err != nil {
				return values.Value{}, err
			}

			if err := fn.checkArity(argCount); err != nil {
				span := chunk.SpanAt(opStart)

				return values.Value{}, locate(err, span.Start, span.End)
			}

			frame.ip = ip
			frame = vm.pushFrame(fn, calleeIdx)
			chunk = frame.chunk
			code = chunk.Code
			ip = 0

		case bytecode.OpReturn:
			retval := vm.discard()
			returning := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]

			vm.forStack = vm.forStack[:returning.forDepth]

			if returning.localsMark != noLocalsMark {
				vm.localsTop = returning.localsMark
			}

			// A return from inside an unfinished expression would
			// otherwise leave that expressions operands behind.
			vm.stack = vm.stack[:returning.stackBase]

			if len(vm.frames) == 0 {
				return retval, nil
			}

			vm.push(retval)

			frame = &vm.frames[len(vm.frames)-1]
			chunk = frame.chunk
			code = chunk.Code
			ip = frame.ip

		case bytecode.OpForBegin:
			step := vm.discard()
			end := vm.discard()
			start := vm.discard()

			state, err := newForState(start, end, step)

			if err != nil {
				span := chunk.SpanAt(opStart)

				return values.Value{}, locate(err, span.Start, span.End)
			}

			vm.forStack = append(vm.forStack, state)

		case bytecode.OpForTest:
			target := bytecode.ReadU32(code, ip)
			ip += 4

			state := &vm.forStack[len(vm.forStack)-1]

			if !state.inRange() {
				ip = int(target)

				break
			}

			counter, convErr := state.counter()

			if convErr != nil {
				span := chunk.SpanAt(opStart)

				return values.Value{}, errors.NewRTError(span.Start, span.End, convErr.Error())
			}

			vm.push(counter)

		case bytecode.OpForNext:
			target := bytecode.ReadU32(code, ip)
			ip += 4

			vm.forStack[len(vm.forStack)-1].advance()

			ip = int(target)

		case bytecode.OpForEnd:
			vm.forStack = vm.forStack[:len(vm.forStack)-1]

		case bytecode.OpBuildList:
			count := int(bytecode.ReadU32(code, ip))
			ip += 4

			elements := make([]values.Value, count)
			copy(elements, vm.stack[len(vm.stack)-count:])
			vm.stack = vm.stack[:len(vm.stack)-count]

			vm.push(values.NewList(elements))

		case bytecode.OpBuildBytes:
			count := int(bytecode.ReadU32(code, ip))
			ip += 4

			data, err := collections.BuildBytes(vm.stack[len(vm.stack)-count:])

			if err != nil {
				span := chunk.SpanAt(opStart)

				return values.Value{}, locate(err, span.Start, span.End)
			}

			vm.stack = vm.stack[:len(vm.stack)-count]
			vm.push(values.NewBytes(data))

		case bytecode.OpBuildMap:
			pairs := int(bytecode.ReadU32(code, ip))
			ip += 4

			mapValue, err := collections.BuildMap(vm.stack[len(vm.stack)-2*pairs:])

			if err != nil {
				span := chunk.SpanAt(opStart)

				return values.Value{}, locate(err, span.Start, span.End)
			}

			vm.stack = vm.stack[:len(vm.stack)-2*pairs]
			vm.push(mapValue)

		case bytecode.OpIndexGet:
			index := vm.discard()
			collection := vm.discard()

			result, err := collections.Get(collection, index)

			if err != nil {
				return values.Value{}, locateIndexFault(err, chunk, opStart)
			}

			vm.push(result)

		case bytecode.OpIndexSet:
			value := vm.discard()
			index := vm.discard()
			collection := vm.discard()

			result, err := collections.Set(collection, index, value)

			if err != nil {
				return values.Value{}, locateIndexFault(err, chunk, opStart)
			}

			vm.push(result)

		case bytecode.OpNeg:
			err = vm.unary(vm.peek().Negated())

		case bytecode.OpNot:
			err = vm.unary(vm.peek().Notted())
		case bytecode.OpBNot:
			err = vm.unary(vm.peek().BNotted())

		case bytecode.OpAdd:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				if result, overflow := values.AddInt64(left, right); !overflow {
					err = vm.binary(values.Int(result), nil)

					break
				}
			}

			err = vm.binary(vm.left().AddedTo(vm.right()))
		case bytecode.OpSub:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				if result, overflow := values.SubInt64(left, right); !overflow {
					err = vm.binary(values.Int(result), nil)

					break
				}
			}

			err = vm.binary(vm.left().SubbedBy(vm.right()))
		case bytecode.OpMul:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				if result, overflow := values.MulInt64(left, right); !overflow {
					err = vm.binary(values.Int(result), nil)

					break
				}
			}

			err = vm.binary(vm.left().MultedBy(vm.right()))
		case bytecode.OpDiv:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole && right != 0 && left%right == 0 {
				if !(left == math.MinInt64 && right == -1) {
					err = vm.binary(values.Int(left/right), nil)

					break
				}
			}

			err = vm.binary(vm.left().DivedBy(vm.right()))
		case bytecode.OpMod:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole && right != 0 && right != -1 {
				err = vm.binary(values.Int(left%right), nil)

				break
			}

			err = vm.binary(vm.left().ModdedBy(vm.right()))
		case bytecode.OpPow:
			err = vm.binary(vm.left().PowedBy(vm.right()))
		case bytecode.OpLShift:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole && right >= 0 {
				err = vm.binary(values.Int(left<<uint64(right)), nil)

				break
			}

			err = vm.binary(vm.left().LShiftedBy(vm.right()))
		case bytecode.OpRShift:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole && right >= 0 {
				err = vm.binary(values.Int(left>>uint64(right)), nil)

				break
			}

			err = vm.binary(vm.left().RShiftedBy(vm.right()))
		case bytecode.OpEq:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Bool(left == right), nil)

				break
			}

			err = vm.binary(vm.left().GetComparisonEe(vm.right()))
		case bytecode.OpNe:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Bool(left != right), nil)

				break
			}

			err = vm.binary(vm.left().GetComparisonNe(vm.right()))
		case bytecode.OpLt:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Bool(left < right), nil)

				break
			}

			err = vm.binary(vm.left().GetComparisonLt(vm.right()))
		case bytecode.OpGt:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Bool(left > right), nil)

				break
			}

			err = vm.binary(vm.left().GetComparisonGt(vm.right()))
		case bytecode.OpLte:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Bool(left <= right), nil)

				break
			}

			err = vm.binary(vm.left().GetComparisonLte(vm.right()))
		case bytecode.OpGte:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Bool(left >= right), nil)

				break
			}

			err = vm.binary(vm.left().GetComparisonGte(vm.right()))
		case bytecode.OpXor:
			err = vm.binary(vm.left().XoredBy(vm.right()))
		case bytecode.OpBAnd:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Int(left&right), nil)

				break
			}

			err = vm.binary(vm.left().BAndedBy(vm.right()))
		case bytecode.OpBOr:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Int(left|right), nil)

				break
			}

			err = vm.binary(vm.left().BOredBy(vm.right()))
		case bytecode.OpBXor:
			if left, right, whole := values.AsInts(vm.left(), vm.right()); whole {
				err = vm.binary(values.Int(left^right), nil)

				break
			}

			err = vm.binary(vm.left().BXoredBy(vm.right()))

		default:
			errors.ModenaPanic("unknown opcode %d, get out the geiger counter", op)
		}

		if err != nil {
			span := chunk.SpanAt(opStart)

			return values.Value{}, locate(err, span.Start, span.End)
		}
	}

	if len(vm.stack) != 1 {
		errors.ModenaPanic("stack imbalance, %d values left at halt", len(vm.stack))
	}

	return vm.stack[0], nil
}

func (vm *VM) callBuiltin(calleeIdx, argCount int, chunk *bytecode.Chunk, opStart int) error {
	args := make([]values.Value, argCount)
	copy(args, vm.stack[calleeIdx+1:])

	result := vm.stack[calleeIdx].Execute(args, vm.ctx)

	if result.Error != nil {
		// A builtin that blames a specific argument underlines that
		// argument, using the spans the compiler recorded for the call.
		if argPos := result.FailArg; argPos > 0 {
			argSpans := chunk.CallArgSpansAt(opStart)

			if argPos <= len(argSpans) && argSpans[argPos-1].Start != nil {
				argSpan := argSpans[argPos-1]

				return locate(result.Error, argSpan.Start, argSpan.End)
			}
		}

		span := chunk.SpanAt(opStart)

		return locate(result.Error, span.Start, span.End)
	}

	vm.stack = vm.stack[:calleeIdx]
	vm.push(result.Value)

	return nil
}

func (vm *VM) checkCallDepth(chunk *bytecode.Chunk, opStart int) error {
	if len(vm.frames) < constants.LIMIT_CALL_DEPTH {
		return nil
	}

	span := chunk.SpanAt(opStart)

	return errors.NewRTError(span.Start, span.End, "Maximum recursion depth exceeded")
}

// An index error underlines the part at fault, the 99 in l[99] rather than the
// whole expression.
func locateIndexFault(err error, chunk *bytecode.Chunk, opStart int) error {
	fault, ok := err.(*collections.Fault)

	if !ok {
		return err
	}

	spans := chunk.IndexSpansAt(opStart)
	span := spans.Idx

	switch fault.Part {
	case collections.PartCollection:
		span = spans.Coll
	case collections.PartValue:
		span = spans.Val
	}

	return locate(fault.Err, span.Start, span.End)
}

func (vm *VM) pushFrame(fn *Function, calleeIdx int) *callFrame {
	chunk := fn.template.Chunk
	locals, mark := vm.takeLocals(chunk)
	copy(locals, vm.stack[calleeIdx+1:])
	vm.stack = vm.stack[:calleeIdx]

	vm.frames = append(vm.frames, callFrame{
		closure:    fn,
		chunk:      chunk,
		locals:     locals,
		stackBase:  calleeIdx,
		forDepth:   len(vm.forStack),
		localsMark: mark,
	})

	return &vm.frames[len(vm.frames)-1]
}

// A call whose locals a nested function captures gets its own, since the shared
// run is reused as soon as the call returns.
func (vm *VM) takeLocals(chunk *bytecode.Chunk) ([]values.Value, int) {
	slots := chunk.NumSlots

	if chunk.LocalsCaptured() {
		return make([]values.Value, slots), noLocalsMark
	}

	// Built on the first call, so a vm that never calls anything, such as an
	// actor running a single builtin, never pays for it.
	if vm.locals == nil {
		vm.locals = make([]values.Value, constants.SIZE_LOCALS_SHARED)
	}

	if vm.localsTop+slots > len(vm.locals) {
		return make([]values.Value, slots), noLocalsMark
	}

	mark := vm.localsTop
	vm.localsTop += slots
	locals := vm.locals[mark:vm.localsTop]

	clear(locals)

	return locals, mark
}

// A loop whose start, end and step are all whole numbers counts in whole numbers,
// so even a very large range stays exact. Any other loop counts in floating
// point.
func newForState(start, end, step values.Value) (forState, error) {
	startNum, err := forBound(start, "Start")

	if err != nil {
		return forState{}, err
	}

	endNum, err := forBound(end, "End")

	if err != nil {
		return forState{}, err
	}

	stepNum, err := forBound(step, "Step")

	if err != nil {
		return forState{}, err
	}

	state := forState{}

	if startNum.IsInt() && endNum.IsInt() && stepNum.IsInt() {
		state.isInt = true
		state.intCounter, _ = startNum.AsInt()
		state.intEnd, _ = endNum.AsInt()
		state.intStep, _ = stepNum.AsInt()

		return state, nil
	}

	state.floatCounter = startNum.AsFloat()
	state.floatEnd = endNum.AsFloat()
	state.floatStep = stepNum.AsFloat()

	return state, nil
}

func forBound(value values.Value, which string) (values.Value, error) {
	if !value.IsNumber() {
		return values.Value{}, errors.NewCallError(which + " value must be a number")
	}

	return value, nil
}

func locate(err error, start, end *errors.Position) error {
	if rtErr, ok := err.(*errors.RTError); ok && rtErr.PosStart == nil {
		rtErr.PosStart = start
		rtErr.PosEnd = end
	}

	return err
}

// left and right are the two operands an operator works on, the left one having
// been pushed first. They are read in place, and binary writes the result over
// the left one, so a hot arithmetic op moves nothing on or off the stack beyond
// that single result.
func (vm *VM) left() values.Value {
	if len(vm.stack) < 2 {
		underflow()
	}

	return vm.stack[len(vm.stack)-2]
}

func (vm *VM) right() values.Value {
	return vm.stack[len(vm.stack)-1]
}

// The error an operator returns carries no position. The loop puts the position
// of the instruction on it in one place, once.
func (vm *VM) binary(result values.Value, opErr error) error {
	if opErr != nil {
		return opErr
	}

	top := len(vm.stack) - 1
	vm.stack[top-1] = result
	vm.stack = vm.stack[:top]

	return nil
}

func (vm *VM) unary(result values.Value, opErr error) error {
	if opErr != nil {
		return opErr
	}

	vm.stack[len(vm.stack)-1] = result

	return nil
}

//go:noinline
func underflow() {
	errors.ModenaPanic("stack underflow")
}

func (vm *VM) push(value values.Value) {
	vm.stack = append(vm.stack, value)
}

// A broken chunk fails as a Modena error rather than as a bare Go index panic.
func (vm *VM) discard() values.Value {
	if len(vm.stack) == 0 {
		underflow()
	}

	top := len(vm.stack) - 1
	value := vm.stack[top]
	vm.stack = vm.stack[:top]

	return value
}

func (vm *VM) peek() values.Value {
	if len(vm.stack) == 0 {
		underflow()
	}

	return vm.stack[len(vm.stack)-1]
}
