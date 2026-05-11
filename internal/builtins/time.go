/*
 *
 * RR2 - internal/builtins/time.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"time"
)

func timeFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("time", 0)))
	}

	// Return high-precision timestamp
	timestamp := float64(time.Now().UnixNano()) / 1e9

	return res.Success(values.NewNumber(timestamp).SetContext(ctx))
}
