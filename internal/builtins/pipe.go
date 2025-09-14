/*
 *
 * RR2 - internal/builtins/pipe.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
)

func pipeFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("pipe", 0),
			ctx,
		))
	}

	readFile, writeFile, err := os.Pipe()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	readHandle := shared.GetNextFileHandle()
	writeHandle := shared.GetNextFileHandle()

	shared.StoreFileHandle(readHandle, readFile)
	shared.StoreFileHandle(writeHandle, writeFile)

	// Return list containing [read_handle, write_handle]
	pipeHandles := []values.Value{
		values.NewNumber(float64(readHandle)),
		values.NewNumber(float64(writeHandle)),
	}

	return res.Success(values.NewList(pipeHandles).SetContext(ctx))
}
