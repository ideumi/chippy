/*
 *
 * RR2 - internal/builtins/dread.go
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

func dreadFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("dread", 1, "handle"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("dread", shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handleID := int(handle64)
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	handle, exists := registry.Dirs.Get(handleID)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid directory handle"))
	}

	// Read next directory entry
	entries, readErr := handle.DirFile.Readdir(1)

	if readErr != nil {
		if readErr == io.EOF {
			// End of directory
			return res.Success(values.NewString(constants.STR_OK))
		}

		return res.Success(values.NewString(constants.STR_ERR))
	}

	if len(entries) == 0 {
		return res.Success(values.NewString(constants.STR_OK))
	}

	return res.Success(values.NewString(entries[0].Name()))
}
