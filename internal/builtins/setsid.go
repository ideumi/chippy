/*
 *
 * RR2 - internal/builtins/setsid.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"syscall"
)

func setsidFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("setsid", 0),
			ctx,
		))
	}

	sid, _, errno := syscall.Syscall(syscall.SYS_SETSID, 0, 0, 0)

	if errno != 0 {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewNumber(float64(sid)).SetContext(ctx))
}
