/*
 *
 * Modena - internal/vm/vm.go
 *
 */

package vm

import (
	"chip-go/internal/bytecode"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"fmt"
	"math"
)

// callFrame is one call that is currently running. Its local variables are kept
// here rather than on the value stack, so one captured by a function defined
// inside it keeps the same address for as long as that function holds it.
type callFrame struct {
	closure  *Function
	chunk    *bytecode.Chunk
	ip       int
	locals   []values.Value
	forDepth int
	callSpan bytecode.Span
}

// forState is one for-loop that is currently running. done means the next step
// would take the counter past the largest whole number that fits, so the loop
// must stop.
type forState struct {
	isInt           bool
	done            bool
	i, end, step    int64
	fi, fend, fstep float64
}

func (s *forState) inRange() bool {
	if s.isInt {
		return !s.done && ((s.step >= 0 && s.i <= s.end) || (s.step < 0 && s.i >= s.end))
	}

	return (s.fstep >= 0 && s.fi <= s.fend) || (s.fstep < 0 && s.fi >= s.fend)
}

func (s *forState) counter() (values.Value, error) {
	if s.isInt {
		return values.NewNumber(s.i), nil
	}

	return values.NewNumberFromFloat(s.fi)
}

func (s *forState) advance() {
	if !s.isInt {
		s.fi += s.fstep

		return
	}

	if (s.step > 0 && s.i > math.MaxInt64-s.step) || (s.step < 0 && s.i < math.MinInt64-s.step) {
		s.done = true

		return
	}

	s.i += s.step
}

type VM struct {
	stack    []values.Value
	frames   []callFrame
	forStack []forState
	ctx      values.Ctx
}

func NewVM(chunk *bytecode.Chunk, ctx values.Ctx) *VM {
	vm := &VM{ctx: ctx}
	vm.frames = append(vm.frames, callFrame{chunk: chunk, locals: make([]values.Value, chunk.NumSlots)})

	return vm
}

func (vm *VM) Run() (values.Value, error) {
	frame := &vm.frames[len(vm.frames)-1]
	code := frame.chunk.Code
	ip := 0

	for ip < len(code) {
		op := bytecode.Op(code[ip])
		span := frame.chunk.Spans[ip]
		ip++

		var err error

		switch op {
		case bytecode.OpConstant:
			index := bytecode.ReadU32(code, ip)
			ip += 4

			// The stored constant is copied as it is loaded, so it
			// is never shared with the program or given a new source
			// position.
			value := frame.chunk.Constants[index].Copy()
			value.SetContext(vm.ctx).SetPos(span.Start, span.End)
			vm.push(value)

		case bytecode.OpDiscard:
			vm.discard()

		case bytecode.OpGetGlobal:
			name := frame.chunk.Names[bytecode.ReadU32(code, ip)]
			ip += 4

			value := vm.ctx.SymbolTable.Get(name)

			if value == nil {
				return nil, errors.NewRTError(span.Start, span.End, fmt.Sprintf("'%s' is not defined", name))
			}

			vm.push(value.SetPos(span.Start, span.End))

		case bytecode.OpDefineGlobal:
			name := frame.chunk.Names[bytecode.ReadU32(code, ip)]
			ip += 4

			vm.ctx.SymbolTable.Set(name, vm.peek())

		case bytecode.OpSetGlobal:
			name := frame.chunk.Names[bytecode.ReadU32(code, ip)]
			ip += 4

			if !vm.ctx.SymbolTable.Exists(name) {
				return nil, errors.NewRTError(span.Start, span.End, fmt.Sprintf("'%s' is not defined", name))
			}

			vm.ctx.SymbolTable.SetInScope(name, vm.peek())

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

			if value == nil {
				return nil, errors.NewRTError(span.Start, span.End, fmt.Sprintf("'%s' is not defined", frame.chunk.LocalNames[slot]))
			}

			vm.push(value.SetPos(span.Start, span.End))

		case bytecode.OpSetLocal:
			slot := int(bytecode.ReadU32(code, ip))
			ip += 4

			frame.locals[slot] = vm.peek()

		case bytecode.OpGetUpvalue:
			index := int(bytecode.ReadU32(code, ip))
			ip += 4

			value := *frame.closure.upvalues[index]

			if value == nil {
				return nil, errors.NewRTError(span.Start, span.End, fmt.Sprintf("'%s' is not defined", frame.closure.template.Upvalues[index].Name))
			}

			vm.push(value.SetPos(span.Start, span.End))

		case bytecode.OpSetUpvalue:
			index := int(bytecode.ReadU32(code, ip))
			ip += 4

			*frame.closure.upvalues[index] = vm.peek()

		case bytecode.OpClosure:
			template := frame.chunk.Functions[bytecode.ReadU32(code, ip)]
			ip += 4

			fn := &Function{BaseValue: values.NewBaseValue(), template: template, globals: vm.ctx}
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

			vm.push(fn.SetPos(span.Start, span.End))

		case bytecode.OpCall:
			argCount := int(bytecode.ReadU32(code, ip))
			ip += 4

			calleeIdx := len(vm.stack) - 1 - argCount
			fn, isChippyFunc := vm.stack[calleeIdx].(*Function)

			if !isChippyFunc {
				if err := vm.callBuiltin(calleeIdx, argCount, span); err != nil {
					return nil, err
				}

				break
			}

			if err := vm.checkCallDepth(span); err != nil {
				return nil, err
			}

			if err := fn.checkArity(argCount); err != nil {
				return nil, locate(err, span.Start, span.End)
			}

			frame.ip = ip
			frame = vm.pushFrame(fn, calleeIdx, span)
			code = frame.chunk.Code
			ip = 0

		case bytecode.OpReturn:
			retval := vm.discard()
			returning := vm.frames[len(vm.frames)-1]
			vm.frames = vm.frames[:len(vm.frames)-1]

			vm.forStack = vm.forStack[:returning.forDepth]

			if len(vm.frames) == 0 {
				return retval, nil
			}

			// The returned value is given the position of the call,
			// so a later error points at the call and not at somewhere
			// inside the function. A call made from outside returns
			// above this and keeps the position it had in the body.
			vm.push(retval.SetPos(returning.callSpan.Start, returning.callSpan.End))

			frame = &vm.frames[len(vm.frames)-1]
			code = frame.chunk.Code
			ip = frame.ip

		case bytecode.OpForBegin:
			step := vm.discard()
			end := vm.discard()
			start := vm.discard()

			state, err := newForState(start, end, step)

			if err != nil {
				return nil, locate(err, span.Start, span.End)
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
				return nil, errors.NewRTError(span.Start, span.End, convErr.Error())
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

			vm.push(values.NewList(elements).SetContext(vm.ctx).SetPos(span.Start, span.End))

		case bytecode.OpBuildBytes:
			count := int(bytecode.ReadU32(code, ip))
			ip += 4

			data, err := bytesFromStack(vm.stack[len(vm.stack)-count:], span)

			if err != nil {
				return nil, err
			}

			vm.stack = vm.stack[:len(vm.stack)-count]
			vm.push(values.NewBytes(data).SetContext(vm.ctx).SetPos(span.Start, span.End))

		case bytecode.OpBuildMap:
			pairs := int(bytecode.ReadU32(code, ip))
			ip += 4

			mapValue, err := mapFromStack(vm.stack[len(vm.stack)-2*pairs:], vm.ctx)

			if err != nil {
				return nil, err
			}

			vm.stack = vm.stack[:len(vm.stack)-2*pairs]
			vm.push(mapValue.SetPos(span.Start, span.End))

		case bytecode.OpIndexGet:
			index := vm.discard()
			collection := vm.discard()

			// An error should underline the part at fault, the 99 in
			// l[99] rather than the whole thing, so each part has
			// its own position.
			spans := frame.chunk.IndexSpansAt(ip - 1)
			result, err := indexGet(collection, index, vm.ctx, spans.Coll, spans.Idx)

			if err != nil {
				return nil, err
			}

			vm.push(result)

		case bytecode.OpIndexSet:
			value := vm.discard()
			index := vm.discard()
			collection := vm.discard()

			spans := frame.chunk.IndexSpansAt(ip - 1)
			result, err := indexSet(collection, index, value, vm.ctx, spans.Coll, spans.Idx, spans.Val)

			if err != nil {
				return nil, err
			}

			vm.push(result)

		case bytecode.OpNeg:
			err = vm.unaryOp(span, values.Value.Negated)

		case bytecode.OpPos:
			vm.push(vm.discard().SetPos(span.Start, span.End))

		// Each takes its operands off the top and leaves one result. The
		// work is handed to the shared values.Value method, so what an
		// operator means lives in one place.
		case bytecode.OpNot:
			err = vm.unaryOp(span, values.Value.Notted)
		case bytecode.OpBNot:
			err = vm.unaryOp(span, values.Value.BNotted)

		case bytecode.OpAdd:
			err = vm.binaryOp(span, values.Value.AddedTo)
		case bytecode.OpSub:
			err = vm.binaryOp(span, values.Value.SubbedBy)
		case bytecode.OpMul:
			err = vm.binaryOp(span, values.Value.MultedBy)
		case bytecode.OpDiv:
			err = vm.binaryOp(span, values.Value.DivedBy)
		case bytecode.OpMod:
			err = vm.binaryOp(span, values.Value.ModdedBy)
		case bytecode.OpPow:
			err = vm.binaryOp(span, values.Value.PowedBy)
		case bytecode.OpLShift:
			err = vm.binaryOp(span, values.Value.LShiftedBy)
		case bytecode.OpRShift:
			err = vm.binaryOp(span, values.Value.RShiftedBy)
		case bytecode.OpEq:
			err = vm.binaryOp(span, values.Value.GetComparisonEe)
		case bytecode.OpNe:
			err = vm.binaryOp(span, values.Value.GetComparisonNe)
		case bytecode.OpLt:
			err = vm.binaryOp(span, values.Value.GetComparisonLt)
		case bytecode.OpGt:
			err = vm.binaryOp(span, values.Value.GetComparisonGt)
		case bytecode.OpLte:
			err = vm.binaryOp(span, values.Value.GetComparisonLte)
		case bytecode.OpGte:
			err = vm.binaryOp(span, values.Value.GetComparisonGte)
		case bytecode.OpXor:
			err = vm.binaryOp(span, values.Value.XoredBy)
		case bytecode.OpBAnd:
			err = vm.binaryOp(span, values.Value.BAndedBy)
		case bytecode.OpBOr:
			err = vm.binaryOp(span, values.Value.BOredBy)
		case bytecode.OpBXor:
			err = vm.binaryOp(span, values.Value.BXoredBy)

		default:
			bytecode.ModenaPanic("unknown opcode %d, get out the geiger counter", op)
		}

		if err != nil {
			return nil, err
		}
	}

	// The program must leave exactly one value. Anything else is a bug.
	if len(vm.stack) != 1 {
		bytecode.ModenaPanic("stack imbalance, %d values left at halt", len(vm.stack))
	}

	return vm.stack[0], nil
}

func (vm *VM) callBuiltin(calleeIdx, argCount int, span bytecode.Span) error {
	args := make([]values.Value, argCount)
	copy(args, vm.stack[calleeIdx+1:])

	result := vm.stack[calleeIdx].Execute(args)

	if result.Error != nil {
		if argPos := result.FailArg; argPos > 0 && argPos <= len(args) {
			if start, end := args[argPos-1].GetPos(); start != nil {
				return locate(result.Error, start, end)
			}
		}

		return locate(result.Error, span.Start, span.End)
	}

	vm.stack = vm.stack[:calleeIdx]
	vm.push(result.Value.SetPos(span.Start, span.End))

	return nil
}

func (vm *VM) checkCallDepth(span bytecode.Span) error {
	if len(vm.frames) >= constants.LIMIT_CALL_DEPTH {
		return errors.NewRTError(span.Start, span.End, "Maximum recursion depth exceeded")
	}

	return nil
}

func (vm *VM) pushFrame(fn *Function, calleeIdx int, span bytecode.Span) *callFrame {
	locals := make([]values.Value, fn.template.Chunk.NumSlots)
	copy(locals, vm.stack[calleeIdx+1:])
	vm.stack = vm.stack[:calleeIdx]

	vm.frames = append(vm.frames, callFrame{
		closure:  fn,
		chunk:    fn.template.Chunk,
		locals:   locals,
		forDepth: len(vm.forStack),
		callSpan: span,
	})

	return &vm.frames[len(vm.frames)-1]
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
		state.i, _ = startNum.AsInt()
		state.end, _ = endNum.AsInt()
		state.step, _ = stepNum.AsInt()

		return state, nil
	}

	state.fi = startNum.AsFloat()
	state.fend = endNum.AsFloat()
	state.fstep = stepNum.AsFloat()

	return state, nil
}

func forBound(value values.Value, which string) (*values.Number, error) {
	num, ok := value.(*values.Number)

	if !ok {
		posStart, posEnd := value.GetPos()

		return nil, errors.NewRTError(posStart, posEnd, which+" value must be a number")
	}

	return num, nil
}

func locate(err error, start, end *errors.Position) error {
	if rtErr, ok := err.(*errors.RTError); ok && rtErr.PosStart == nil {
		rtErr.PosStart = start
		rtErr.PosEnd = end
	}

	return err
}

// binaryOp takes the right operand off the stack first and the left one second,
// because that is the order they were put there.
func (vm *VM) binaryOp(span bytecode.Span, op func(values.Value, values.Value) (values.Value, error)) error {
	right := vm.discard()
	left := vm.discard()

	result, err := op(left, right)

	if err != nil {
		return locate(err, span.Start, span.End)
	}

	vm.push(result.SetPos(span.Start, span.End))

	return nil
}

func (vm *VM) unaryOp(span bytecode.Span, op func(values.Value) (values.Value, error)) error {
	result, err := op(vm.discard())

	if err != nil {
		return locate(err, span.Start, span.End)
	}

	vm.push(result.SetPos(span.Start, span.End))

	return nil
}

func (vm *VM) push(value values.Value) {
	vm.stack = append(vm.stack, value)
}

// A broken chunk fails as a Modena error rather than as a bare Go index panic.
func (vm *VM) discard() values.Value {
	if len(vm.stack) == 0 {
		bytecode.ModenaPanic("stack underflow")
	}

	top := len(vm.stack) - 1
	value := vm.stack[top]
	vm.stack = vm.stack[:top]

	return value
}

func (vm *VM) peek() values.Value {
	if len(vm.stack) == 0 {
		bytecode.ModenaPanic("stack underflow")
	}

	return vm.stack[len(vm.stack)-1]
}
