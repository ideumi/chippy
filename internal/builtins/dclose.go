/*
 *
 * RR2 - internal/builtins/dclose.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func dcloseFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("dclose", 1, "handle"))
	}

	handleNum := args[0]
	if !handleNum.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("dclose", shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handleID := int(handle64)
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	handle, exists := registry.Dirs.Extract(handleID)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid directory handle"))
	}

	registry.Alloc.Free(handleID)

	err = handle.DirFile.Close()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(constants.STR_OK))
}
