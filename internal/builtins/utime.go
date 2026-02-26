/*
 *
 * RR2 - internal/builtins/utime.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"math"
	"syscall"
)

func utimeFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("utime", 3, "path, atime, mtime"),
			ctx,
		))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("utime", shared.PositionFirst, shared.TypeString, "path"),
			ctx,
		))
	}

	atimeNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("utime", shared.PositionSecond, shared.TypeNumber, "atime"),
			ctx,
		))
	}

	mtimeNum, ok := args[2].(*values.Number)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("utime", shared.PositionThird, shared.TypeNumber, "mtime"),
			ctx,
		))
	}

	toTimespec := func(ts float64) syscall.Timespec {
		sec := int64(math.Floor(ts))
		nsec := int64((ts - float64(sec)) * 1e9)
		return syscall.Timespec{Sec: sec, Nsec: nsec}
	}

	ts := []syscall.Timespec{
		toTimespec(atimeNum.Value),
		toTimespec(mtimeNum.Value),
	}

	err := syscall.UtimesNano(pathStr.Value, ts)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
