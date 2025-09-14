/*
 *
 * RR2 - internal/builtins/waitpid.go
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

func waitpidFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("waitpid", 1, "pid"),
			ctx,
		))
	}

	pidNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("waitpid", shared.TypeNumber, "pid"),
			ctx,
		))
	}

	targetPid := int(pidNum.Value)

	var status syscall.WaitStatus
	pid, err := syscall.Wait4(targetPid, &status, 0, nil)

	// No processes or error
	if err != nil || pid != targetPid {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Return exit status of specific child process
	return res.Success(values.NewNumber(float64(status.ExitStatus())).SetContext(ctx))
}
