/*
 *
 * RR2 - internal/builtins/fwrite.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"io"
)

func fwriteFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("fwrite", 2, "bytes, handle"),
			ctx,
		))
	}

	bytesVal, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("fwrite", shared.PositionFirst, shared.TypeBytes, shared.TypeBytes),
			ctx,
		))
	}

	handleNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("fwrite", shared.PositionSecond, shared.TypeNumber, "handle"),
			ctx,
		))
	}

	handle := int(handleNum.Value)

	registry := orchestrator.Get().GetRegistry(ctx)
	file, fileExists := registry.Files.Get(handle)
	procHandle, procExists := registry.Processes.Get(handle)

	var writer io.Writer

	if fileExists {
		writer = file
	} else if procExists && procHandle.Stdin != nil {
		writer = procHandle.Stdin
	} else {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid handle"),
			ctx,
		))
	}

	if len(bytesVal.Data) == 0 {
		return res.Success(values.NewNumber(constants.NUM_NUL).SetContext(ctx))
	}

	n, err := writer.Write(bytesVal.Data)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewNumber(float64(n)).SetContext(ctx))
}
