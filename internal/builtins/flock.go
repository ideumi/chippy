/*
 *
 * RR2 - internal/builtins/flock.go
 *
 */

// NOTE: This is UNIX only for now..

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"syscall"
)

const (
	LOCK_SH = 1 // Shared lock
	LOCK_EX = 2 // Exclusive lock
	LOCK_NB = 4 // Non-blocking (can be ORed with SH/EX)
	LOCK_UN = 8 // Unlock
)

func flockFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position
		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("flock", 2, "file handle, operation"),
			ctx,
		))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("flock", shared.PositionFirst, shared.TypeNumber, "file handle"),
			ctx,
		))
	}

	opNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("flock", shared.PositionSecond, shared.TypeNumber, "operation"),
			ctx,
		))
	}

	handleID := int(handleNum.Value)

	file, exists := shared.GetFileHandle(handleID)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid handle"),
			ctx,
		))
	}

	fd := int(file.Fd())
	operation := int(opNum.Value)

	// Thermonuclear flock();
	_, _, errno := syscall.Syscall(syscall.SYS_FLOCK, uintptr(fd), uintptr(operation), 0)

	if errno != 0 {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
