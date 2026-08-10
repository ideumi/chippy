/*
 *
 * Chippy - internal/builtins/len.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func lenFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("len", 1, "value"))
	}

	value := args[0]

	if str, ok := values.AsString(value); ok {
		return res.Success(values.NewNumber(str.RuneCount()))
	}

	if list, ok := values.AsList(value); ok {
		return res.Success(values.NewNumber(len(list.Elements)))
	}

	if bytesVal, ok := values.AsBytes(value); ok {
		return res.Success(values.NewNumber(len(bytesVal.Data)))
	}

	if mapVal, ok := values.AsMap(value); ok {
		return res.Success(values.NewNumber(len(mapVal.Keys)))
	}

	return res.FailAt(1,
		shared.Errors.InvalidValue("len() can only be used on strings, lists, bytes, and maps"))
}
