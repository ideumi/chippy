/*
 *
 * RR2 - internal/builtins/fwrite.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"io"
)

func fwriteFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("fwrite", 2, "bytes, handle"))
	}

	bytesVal, ok := args[0].(*values.Bytes)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("fwrite", shared.PositionFirst, shared.TypeBytes, shared.TypeBytes))
	}

	handleNum, ok := args[1].(*values.Number)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("fwrite", shared.PositionSecond, shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	file, fileExists := registry.Files.Get(handle)
	procHandle, procExists := registry.Processes.Get(handle)

	var writer io.Writer

	if fileExists {
		writer = file
	} else if procExists && procHandle.Stdin != nil {
		writer = procHandle.Stdin
	} else {
		return res.FailAt(2, shared.Errors.InvalidValue("Invalid handle"))
	}

	if len(bytesVal.Data) == 0 {
		return res.Success(values.NewNumber(constants.NUM_NUL).SetContext(ctx))
	}

	n, err := writer.Write(bytesVal.Data)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewNumber(n).SetContext(ctx))
}
