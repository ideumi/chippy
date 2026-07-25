/*
 *
 * RR2 - internal/values/builtin.go
 *
 */

package values

import (
	"chip-go/internal/errors"
	"fmt"
)

type BuiltInFunction struct {
	*BaseValue
	Name string
	Fn   func([]Value, Ctx) *RuntimeResult
}

func NewBuiltInFunction(name string, fn func([]Value, Ctx) *RuntimeResult) *BuiltInFunction {
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

func (bf *BuiltInFunction) SetContext(ctx Ctx) Value {
	bf.BaseValue.SetContext(ctx)
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
		return res.Fail("Built-in function context is nil")
	}

	return bf.Fn(args, bf.context)
}
