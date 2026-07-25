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

	switch v := value.(type) {
	case *values.String:
		return res.Success(values.NewNumber(utf8.RuneCountInString(v.Value)).SetContext(ctx))
	case *values.List:
		return res.Success(values.NewNumber(len(v.Elements)).SetContext(ctx))
	case *values.Bytes:
		return res.Success(values.NewNumber(len(v.Data)).SetContext(ctx))
	case *values.Map:
		return res.Success(values.NewNumber(len(v.Keys)).SetContext(ctx))
	default:
		return res.FailAt(1,
			shared.Errors.InvalidValue("len() can only be used on strings, lists, bytes, and maps"))
	}
}
