/*
 *
 * RR2 - internal/builtins/fsync.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func fsyncFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("fsync", 1, "handle"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("fsync", shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	file, exists := registry.Files.Get(handle)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid handle"))
	}

	err = file.Sync()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(constants.STR_OK))
}
