/*
 *
 * RR2 - internal/builtins/lenv.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"chip-go/thirdparty/runewidth"
)

func lenvFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("lenv", 1, "string"))
	}

	stringArg, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("lenv", shared.TypeString, "string"))
	}

	str := stringArg.Value
	width := runewidth.StringWidth(str)

	return res.Success(values.NewNumber(width))
}
