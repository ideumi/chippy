/*
 *
 * RR2 - internal/builtins/kill.go
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

func killFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("kill", 2, "pid, signal")))
	}

	pidNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("kill", shared.PositionFirst, shared.TypeNumber, "pid")))
	}

	signalNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("kill", shared.PositionSecond, shared.TypeNumber, "signal")))
	}

	pid64, err := pidNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	sig64, err := signalNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	pid := int(pid64)
	signal := syscall.Signal(sig64)

	err = syscall.Kill(pid, signal)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Success
	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
