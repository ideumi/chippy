/*
 *
 * Chippy - internal/values/function.go
 *
 */

package values

import (
	"chip-go/internal/errors"
	"fmt"
)

type Callable interface {
	ArgCount() int
	CallableName() string
}

type BoundaryClosure interface {
	Copy() Value
	TransferCells() []*Value
	SetTransferCells(cells []*Value)
	RebindGlobals(globals Ctx)
}

type BuiltInFunction struct {
	OperatorDefaults
	Name string
	Fn   func([]Value, Ctx) RuntimeResult
}

func NewBuiltInFunction(name string, fn func([]Value, Ctx) RuntimeResult) Value {
	return fromHeap(TagFunc, &BuiltInFunction{Name: name, Fn: fn})
}

func (bf *BuiltInFunction) String() string {
	return fmt.Sprintf("<built-in function %s>", bf.Name)
}

func (bf *BuiltInFunction) Copy() Value {
	return fromHeap(TagFunc, &BuiltInFunction{Name: bf.Name, Fn: bf.Fn})
}

func (bf *BuiltInFunction) IsTrue() bool {
	return true
}

func (bf *BuiltInFunction) Execute(args []Value, ctx Ctx) RuntimeResult {
	if ctx == nil {
		return NewRuntimeResult().Fail("Built-in function context is nil")
	}

	return bf.Fn(args, ctx)
}

// FailArg is which argument a builtin blames, counted from one, or zero when it
// blames the call as a whole.
type RuntimeResult struct {
	Value   Value
	Error   error
	FailArg int
}

func NewRuntimeResult() RuntimeResult {
	return RuntimeResult{}
}

func (rr RuntimeResult) Success(value Value) RuntimeResult {
	return RuntimeResult{Value: value}
}

func (rr RuntimeResult) Failure(err error) RuntimeResult {
	return RuntimeResult{Error: err}
}

func (rr RuntimeResult) Fail(details string) RuntimeResult {
	return rr.Failure(errors.NewCallError(details))
}

func (rr RuntimeResult) FailAt(arg int, details string) RuntimeResult {
	return RuntimeResult{Error: errors.NewCallError(details), FailArg: arg}
}
