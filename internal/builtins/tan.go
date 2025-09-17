/*
 *
 * RR2 - internal/builtins/tan.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"math"
)

func tanFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("tan", 1, "radians"),
			ctx,
		))
	}

	radiansArg, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("tan", shared.TypeNumber, "radians"),
			ctx,
		))
	}

	radians := radiansArg.Value
	result := math.Tan(radians)

	return res.Success(values.NewNumber(result).SetContext(ctx))
}
