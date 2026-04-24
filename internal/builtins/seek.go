/*
 *
 * RR2 - internal/builtins/seek.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func seekFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("seek", 3, "handle, offset, whence"),
			ctx,
		))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("seek", shared.PositionFirst, shared.TypeNumber, "handle"),
			ctx,
		))
	}

	offsetNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("seek", shared.PositionSecond, shared.TypeNumber, "offset"),
			ctx,
		))
	}

	whenceNum, ok := args[2].(*values.Number)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("seek", shared.PositionThird, shared.TypeNumber, "whence"),
			ctx,
		))
	}

	handle := int(handleNum.Value)
	offset := int64(offsetNum.Value)
	whence := int(whenceNum.Value)

	// Validate whence parameter
	if whence < 0 || whence > 2 {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("whence must be 0 (start), 1 (current), or 2 (end)"),
			ctx,
		))
	}

	registry := orchestrator.Get().GetRegistry(ctx)
	file, exists := registry.Files.Get(handle)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid handle"),
			ctx,
		))
	}

	newPos, err := file.Seek(offset, whence)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewNumber(float64(newPos)).SetContext(ctx))
}
