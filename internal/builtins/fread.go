/*
 *
 * RR2 - internal/builtins/fread.go
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

func freadFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("fread", 2, "handle, count"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("fread", shared.PositionFirst, shared.TypeNumber, "handle"))
	}

	countNum := args[1]

	if !countNum.IsNumber() {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("fread", shared.PositionSecond, shared.TypeNumber, "count"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	count64, err := countNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)
	count := int(count64)

	if count < 0 {
		return res.FailAt(2, shared.Errors.InvalidValue("Count must be non-negative"))
	}

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	file, fileExists := registry.Files.Get(handle)
	procHandle, procExists := registry.Processes.Get(handle)

	var reader io.Reader

	if fileExists {
		reader = file
	} else if procExists && procHandle.Stdout != nil {
		reader = procHandle.Stdout
	} else {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid handle"))
	}

	var buffer []byte
	var bytesRead int

	if count == 0 {
		buffer, err = io.ReadAll(reader)
		bytesRead = len(buffer)
	} else {
		buffer = make([]byte, count)
		bytesRead, err = io.ReadFull(reader, buffer)
	}

	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewBytes(buffer[:bytesRead]))
}
