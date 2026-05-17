/*
 *
 * RR2 - internal/builtins/cos.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"math"
)

func cosFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("cos", 1, "radians")))
	}

	radiansArg, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("cos", shared.TypeNumber, "radians")))
	}

	radians := radiansArg.AsFloat()

	num, err := values.NewNumberFromFloat(math.Cos(radians))

	if err != nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(posStart, posEnd, err.Error()))
	}

	return res.Success(num.SetContext(ctx))
}
