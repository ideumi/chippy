/*
 *
 * RR2 - internal/builtins/utime.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"math"
	"syscall"
)

func utimeFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("utime", 3, "path, atime, mtime"))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("utime", shared.PositionFirst, shared.TypeString, "path"))
	}

	atimeNum, ok := args[1].(*values.Number)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("utime", shared.PositionSecond, shared.TypeNumber, "atime"))
	}

	mtimeNum, ok := args[2].(*values.Number)

	if !ok {
		return res.FailAt(3,
			shared.Errors.InvalidArgTypePositionalWithHint("utime", shared.PositionThird, shared.TypeNumber, "mtime"))
	}

	toTimespec := func(ts float64) syscall.Timespec {
		sec := int64(math.Floor(ts))
		nsec := int64((ts - float64(sec)) * 1e9)
		return syscall.Timespec{Sec: sec, Nsec: nsec}
	}

	ts := []syscall.Timespec{
		toTimespec(atimeNum.AsFloat()),
		toTimespec(mtimeNum.AsFloat()),
	}

	err := syscall.UtimesNano(pathStr.Value, ts)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
