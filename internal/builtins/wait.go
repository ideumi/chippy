/*
 *
 * RR2 - internal/builtins/wait.go
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

func waitFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("wait", 0),
			ctx,
		))
	}

	var status syscall.WaitStatus

	_, err := syscall.Wait4(-1, &status, 0, nil)

	// No child processes or error
	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Exit status of child process
	return res.Success(values.NewNumber(float64(status.ExitStatus())).SetContext(ctx))
}
