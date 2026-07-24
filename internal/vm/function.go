/*
 *
 * Modena - internal/vm/function.go
 *
 */

package vm

import (
	"chip-go/internal/bytecode"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"fmt"
)

type Function struct {
	*values.BaseValue
	template *bytecode.FunctionTemplate
	globals  values.Ctx
	upvalues []*values.Value
}

func (f *Function) String() string {
	if f.template.Name != "" {
		return fmt.Sprintf("<function %s>", f.template.Name)
	}

	return constants.BASE_FUNC_NAME_FN_ANON
}

func (f *Function) SetPos(posStart, posEnd *errors.Position) values.Value {
	f.BaseValue.SetPos(posStart, posEnd)
	return f
}

func (f *Function) SetContext(ctx values.Ctx) values.Value {
	f.BaseValue.SetContext(ctx)
	return f
}

func (f *Function) Copy() values.Value {
	start, end := f.GetPos()

	clone := &Function{
		BaseValue: values.NewBaseValue(),
		template:  f.template,
		globals:   f.globals,
		upvalues:  f.upvalues,
	}
	clone.SetPos(start, end)
	clone.SetContext(f.GetContext())

	return clone
}

func (f *Function) IsTrue() bool {
	return true
}

func (f *Function) ArgCount() int {
	return f.template.Arity
}

func (f *Function) CallableName() string {
	return f.template.Name
}

func (f *Function) TransferCells() []*values.Value {
	return f.upvalues
}

func (f *Function) SetTransferCells(cells []*values.Value) {
	f.upvalues = cells
}

func (f *Function) RebindGlobals(globals values.Ctx) {
	f.globals = globals
}

func (f *Function) Execute(args []values.Value) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if err := f.checkArity(len(args)); err != nil {
		return res.Failure(err)
	}

	vm := NewVM(f.template.Chunk, f.globals)
	frame := &vm.frames[0]
	frame.closure = f
	copy(frame.locals, args)

	value, err := vm.Run()

	if err != nil {
		return res.Failure(err)
	}

	return res.Success(value)
}

func (f *Function) checkArity(got int) error {
	if got == f.template.Arity {
		return nil
	}

	start, end := f.GetPos()
	which := "Too few"

	if got > f.template.Arity {
		which = "Too many"
	}

	return errors.NewRTError(
		start, end,
		fmt.Sprintf("%s args passed into '%s'. Expected %d, got %d", which, f.template.Name, f.template.Arity, got))
}
