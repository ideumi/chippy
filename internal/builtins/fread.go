/*
 *
 * RR2 - internal/builtins/fread.go
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

func freadFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("fread", 2, "handle, count"),
			ctx,
		))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("fread", shared.PositionFirst, shared.TypeNumber, "handle"),
			ctx,
		))
	}

	countNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("fread", shared.PositionSecond, shared.TypeNumber, "count"),
			ctx,
		))
	}

	handle := int(handleNum.Value)

	count := int(countNum.Value)

	if count < 0 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Count must be non-negative"),
			ctx,
		))
	}

	registry := orchestrator.Get().GetRegistry(ctx)
	file, fileExists := registry.Files.Get(handle)
	procHandle, procExists := registry.Processes.Get(handle)

	var reader io.Reader

	if fileExists {
		reader = file
	} else if procExists && procHandle.Stdout != nil {
		reader = procHandle.Stdout
	} else {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid handle"),
			ctx,
		))
	}

	var buffer []byte
	var n int
	var err error

	if count == 0 {
		buffer, err = io.ReadAll(reader)
		n = len(buffer)
	} else {
		buffer = make([]byte, count)
		n, err = io.ReadFull(reader, buffer)
	}

	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewBytes(buffer[:n]).SetContext(ctx))
}
