/*
 *
 * RR2 - internal/builtins/mkdir.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"strconv"
	"syscall"
)

func mkdirFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("mkdir", 2, "path, mode"))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("mkdir", shared.PositionFirst, shared.TypeString, "path"))
	}

	modeNum, ok := args[1].(*values.Number)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("mkdir", shared.PositionSecond, shared.TypeNumber, "mode"))
	}

	mode64, err := modeNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	// Convert octal notation to proper file mode
	modeStr := strconv.FormatInt(mode64, 10)
	octalMode, err := strconv.ParseInt(modeStr, 8, 32)

	if err != nil {
		return res.FailAt(2, "Invalid octal mode: "+modeStr)
	}

	err = syscall.Mkdir(pathStr.Value, uint32(octalMode))

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
