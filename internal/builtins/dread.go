/*
 *
 * RR2 - internal/builtins/dread.go
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

func dreadFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("dread", 1, "handle")))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("dread", shared.TypeNumber, "handle")))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handleID := int(handle64)
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	handle, exists := registry.Dirs.Get(handleID)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid directory handle")))
	}

	// Read next directory entry
	entries, readErr := handle.DirFile.Readdir(1)

	if readErr != nil {
		if readErr == io.EOF {
			// End of directory
			return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
		}

		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	if len(entries) == 0 {
		return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
	}

	return res.Success(values.NewString(entries[0].Name()).SetContext(ctx))
}
