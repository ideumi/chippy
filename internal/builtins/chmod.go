/*
 *
 * RR2 - internal/builtins/chmod.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"strconv"
	"syscall"
)

func chmodFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("chmod", 2, "path, mode"),
			ctx,
		))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("chmod", shared.PositionFirst, shared.TypeString, "path"),
			ctx,
		))
	}

	modeNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("chmod", shared.PositionSecond, shared.TypeNumber, "mode"),
			ctx,
		))
	}

	// Convert octal notation to proper file mode
	modeStr := strconv.FormatInt(int64(modeNum.Value), 10)
	octalMode, err := strconv.ParseInt(modeStr, 8, 32)

	if err != nil {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Invalid octal mode: "+modeStr,
			ctx,
		))
	}

	err = syscall.Chmod(pathStr.Value, uint32(octalMode))

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
