/*
 *
 * Chippy - internal/builtins/fopen.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os"
)

func fopenFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("fopen", 2, "path, mode"))
	}

	pathStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("fopen", shared.PositionFirst, shared.TypeString, "path"))
	}

	modeStr, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("fopen", shared.PositionSecond, shared.TypeString, "mode"))
	}

	path := pathStr.Value
	mode := modeStr.Value

	var flag int

	switch mode {
	case "r":
		flag = os.O_RDONLY
	case "w":
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "a":
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "rw":
		flag = os.O_RDWR | os.O_CREATE
	default:
		return res.FailAt(2, shared.Errors.InvalidValue("Invalid mode. Use 'r', 'w', 'a', or 'rw'"))
	}

	file, err := os.OpenFile(path, flag, 0644)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	handle := registry.Alloc.Alloc()
	registry.Files.Store(handle, file)

	return res.Success(values.NewNumber(handle))
}
