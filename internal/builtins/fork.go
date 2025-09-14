/*
 *
 * RR2 - internal/builtins/fork.go
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

func forkFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("fork", 0),
			ctx,
		))
	}

	pid, _, errno := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)

	if errno != 0 {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Fork succeeded
	// In parent: pid contains child PID
	// In child: pid is 0
	return res.Success(values.NewNumber(float64(pid)).SetContext(ctx))
}
