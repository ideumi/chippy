/*
 *
 * RR2 - internal/values/function.go
 *
 */

package values

import (
	"chip-go/internal/ast"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"fmt"
)

type Function struct {
	*BaseValue
	Name             string
	BodyNode         ast.Node
	ArgNames         []string
	ShouldAutoReturn bool
}

func NewFunction(name string, bodyNode ast.Node, argNames []string, shouldAutoReturn bool) *Function {
	return &Function{
		BaseValue:        NewBaseValue(),
		Name:             name,
		BodyNode:         bodyNode,
		ArgNames:         argNames,
		ShouldAutoReturn: shouldAutoReturn,
	}
}

func (f *Function) String() string {
	if f.Name != "" {
		return fmt.Sprintf("<function %s>", f.Name)
	}

	return fmt.Sprintf(constants.BASE_FUNC_NAME_FN_ANON)
}

func (f *Function) SetPos(posStart, posEnd *errors.Position) Value {
	f.BaseValue.SetPos(posStart, posEnd)
	return f
}

func (f *Function) SetContext(context interface{}) Value {
	f.BaseValue.SetContext(context)
	return f
}

// Shallow by design: f.context is the closure capture, and every copy of the
// same function must share it so assignments to captured variables stay
// visible across calls. Cross-actor isolation is the exception, and is
// handled separately by orchestrator.IsolateForTransfer.
func (f *Function) Copy() Value {
	copy := NewFunction(f.Name, f.BodyNode, f.ArgNames, f.ShouldAutoReturn)
	copy.SetPos(f.posStart, f.posEnd)
	copy.SetContext(f.context)

	return copy
}

func (f *Function) IsTrue() bool {
	return true
}

type InterpreterInterface interface {
	Visit(node ast.Node, ctx interface{}) *RuntimeResult
}

func (f *Function) Execute(args []Value) *RuntimeResult {
	res := NewRuntimeResult()

	if globalInterpreter == nil {
		return res.Failure(errors.NewRTError(
			f.posStart, f.posEnd,
			"Interpreter not available",
			f.context,
		))
	}

	if len(args) > len(f.ArgNames) {
		return res.Failure(errors.NewRTError(
			f.posStart, f.posEnd,
			fmt.Sprintf("Too many args passed into '%s'. Expected %d, got %d", f.Name, len(f.ArgNames), len(args)),
			f.context,
		))
	}

	if len(args) < len(f.ArgNames) {
		return res.Failure(errors.NewRTError(
			f.posStart, f.posEnd,
			fmt.Sprintf("Too few args passed into '%s'. Expected %d, got %d", f.Name, len(f.ArgNames), len(args)),
			f.context,
		))
	}

	execCtx := context.NewContext(f.Name, f.context.(*context.Context), nil)

	for i, argName := range f.ArgNames {
		execCtx.SymbolTable.Set(argName, args[i])
	}

	value := res.Register(globalInterpreter.Visit(f.BodyNode, execCtx))
	if res.ShouldReturn() && res.FuncReturnValue == nil {
		return res
	}

	var returnValue Value
	if res.FuncReturnValue != nil || f.ShouldAutoReturn {
		if res.FuncReturnValue != nil {
			returnValue = res.FuncReturnValue
		} else {
			returnValue = value
		}
	} else {
		returnValue = NewNumber(constants.NUM_NUL)
	}

	return res.Success(returnValue)
}

var globalInterpreter InterpreterInterface

func SetGlobalInterpreter(interpreter InterpreterInterface) {
	globalInterpreter = interpreter
}

type BuiltInFunction struct {
	*BaseValue
	Name string
	Fn   func([]Value, interface{}) *RuntimeResult
}

func NewBuiltInFunction(name string, fn func([]Value, interface{}) *RuntimeResult) *BuiltInFunction {
	return &BuiltInFunction{
		BaseValue: NewBaseValue(),
		Name:      name,
		Fn:        fn,
	}
}

func (bf *BuiltInFunction) String() string {
	return fmt.Sprintf("<built-in function %s>", bf.Name)
}

func (bf *BuiltInFunction) SetPos(posStart, posEnd *errors.Position) Value {
	bf.BaseValue.SetPos(posStart, posEnd)
	return bf
}

func (bf *BuiltInFunction) SetContext(context interface{}) Value {
	bf.BaseValue.SetContext(context)
	return bf
}

func (bf *BuiltInFunction) Copy() Value {
	copy := NewBuiltInFunction(bf.Name, bf.Fn)
	copy.SetPos(bf.posStart, bf.posEnd)
	copy.SetContext(bf.context)

	return copy
}

func (bf *BuiltInFunction) IsTrue() bool {
	return true
}

func (bf *BuiltInFunction) Execute(args []Value) *RuntimeResult {
	res := NewRuntimeResult()

	if bf.context == nil {
		return res.Failure(errors.NewRTError(
			bf.posStart, bf.posEnd,
			"Built-in function context is nil",
			nil,
		))
	}

	return bf.Fn(args, bf.context)
}
