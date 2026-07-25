/*
 *
 * RR2 - internal/builtins/sleep.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"time"
)

func sleepFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sleep", 1, "milliseconds"))
	}

	millisecondsNum, ok := args[0].(*values.Number)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("sleep", shared.TypeNumber, "milliseconds to sleep"))
	}

	milliseconds := millisecondsNum.AsFloat()

	if milliseconds < 0 {
		return res.FailAt(1, shared.Errors.InvalidValue("Sleep duration must be non-negative"))
	}

	duration := time.Duration(milliseconds) * time.Millisecond
	time.Sleep(duration)

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
