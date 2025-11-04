/*
 *
 * RR2 - internal/builtins/lenv.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"chip-go/thirdparty/runewidth"
)

func lenvFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("lenv", 1, "string"),
			ctx,
		))
	}

	stringArg, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("lenv", shared.TypeString, "string"),
			ctx,
		))
	}

	str := stringArg.Value
	width := runewidth.StringWidth(str)

	return res.Success(values.NewNumber(float64(width)).SetContext(ctx))
}
