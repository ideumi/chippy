/*
 *
 * RR2 - internal/builtins/lutime.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"math"

	"golang.org/x/sys/unix"
)

func lutimeFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("lutime", 3, "path, atime, mtime"))
	}

	pathStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("lutime", shared.PositionFirst, shared.TypeString, "path"))
	}

	atimeNum := args[1]

	if !atimeNum.IsNumber() {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("lutime", shared.PositionSecond, shared.TypeNumber, "atime"))
	}

	mtimeNum := args[2]

	if !mtimeNum.IsNumber() {
		return res.FailAt(3,
			shared.Errors.InvalidArgTypePositionalWithHint("lutime", shared.PositionThird, shared.TypeNumber, "mtime"))
	}

	toTimespec := func(ts float64) unix.Timespec {
		sec := int64(math.Floor(ts))
		nsec := int64((ts - float64(sec)) * 1e9)
		return unix.Timespec{Sec: sec, Nsec: nsec}
	}

	ts := []unix.Timespec{
		toTimespec(atimeNum.AsFloat()),
		toTimespec(mtimeNum.AsFloat()),
	}

	err := unix.UtimesNanoAt(unix.AT_FDCWD, pathStr.Value, ts, unix.AT_SYMLINK_NOFOLLOW)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(constants.STR_OK))
}
