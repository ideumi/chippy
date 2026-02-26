/*
 *
 * RR2 - internal/builtins/lutime.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"math"

	"golang.org/x/sys/unix"
)

func lutimeFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("lutime", 3, "path, atime, mtime"),
			ctx,
		))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("lutime", shared.PositionFirst, shared.TypeString, "path"),
			ctx,
		))
	}

	atimeNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("lutime", shared.PositionSecond, shared.TypeNumber, "atime"),
			ctx,
		))
	}

	mtimeNum, ok := args[2].(*values.Number)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("lutime", shared.PositionThird, shared.TypeNumber, "mtime"),
			ctx,
		))
	}

	toTimespec := func(ts float64) unix.Timespec {
		sec := int64(math.Floor(ts))
		nsec := int64((ts - float64(sec)) * 1e9)
		return unix.Timespec{Sec: sec, Nsec: nsec}
	}

	ts := []unix.Timespec{
		toTimespec(atimeNum.Value),
		toTimespec(mtimeNum.Value),
	}

	err := unix.UtimesNanoAt(unix.AT_FDCWD, pathStr.Value, ts, unix.AT_SYMLINK_NOFOLLOW)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
