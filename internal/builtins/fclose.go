/*
 *
 * Chippy - internal/builtins/fclose.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func fcloseFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("fclose", 1, "handle"))
	}

	handleNum := args[0]
	if !handleNum.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("fclose", shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)

	if handle >= 0 && handle <= 2 {
		return res.FailAt(1, shared.Errors.InvalidValue("Cannot close standard handles (0, 1, 2)"))
	}

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	file, exists := registry.Files.Extract(handle)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid handle"))
	}

	registry.Alloc.Free(handle)

	err = file.Close()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(constants.STR_OK))
}
