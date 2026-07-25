/*
 *
 * RR2 - internal/builtins/len.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"unicode/utf8"
)

func lenFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("len", 1, "value"))
	}

	value := args[0]

	switch typed := value.(type) {
	case *values.String:
		return res.Success(values.NewNumber(utf8.RuneCountInString(typed.Value)).SetContext(ctx))
	case *values.List:
		return res.Success(values.NewNumber(len(typed.Elements)).SetContext(ctx))
	case *values.Bytes:
		return res.Success(values.NewNumber(len(typed.Data)).SetContext(ctx))
	case *values.Map:
		return res.Success(values.NewNumber(len(typed.Keys)).SetContext(ctx))
	default:
		return res.FailAt(1,
			shared.Errors.InvalidValue("len() can only be used on strings, lists, bytes, and maps"))
	}
}
