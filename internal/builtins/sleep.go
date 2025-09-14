/*
 *
 * RR2 - internal/builtins/sleep.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"time"
)

func sleepFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("sleep", 1, "milliseconds"),
			ctx,
		))
	}

	millisecondsNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("sleep", shared.TypeNumber, "milliseconds to sleep"),
			ctx,
		))
	}

	milliseconds := millisecondsNum.Value

	if milliseconds < 0 {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Sleep duration must be non-negative"),
			ctx,
		))
	}

	duration := time.Duration(milliseconds) * time.Millisecond
	time.Sleep(duration)

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
