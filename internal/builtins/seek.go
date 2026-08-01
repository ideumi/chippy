/*
 *
 * Chippy - internal/builtins/seek.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func seekFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("seek", 3, "handle, offset, whence"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("seek", shared.PositionFirst, shared.TypeNumber, "handle"))
	}

	offsetNum := args[1]

	if !offsetNum.IsNumber() {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("seek", shared.PositionSecond, shared.TypeNumber, "offset"))
	}

	whenceNum := args[2]

	if !whenceNum.IsNumber() {
		return res.FailAt(3,
			shared.Errors.InvalidArgTypePositionalWithHint("seek", shared.PositionThird, shared.TypeNumber, "whence"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	offset, err := offsetNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	whence64, err := whenceNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)
	whence := int(whence64)

	// Validate whence parameter
	if whence < 0 || whence > 2 {
		return res.FailAt(3,
			shared.Errors.InvalidValue("whence must be 0 (start), 1 (current), or 2 (end)"))
	}

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	file, exists := registry.Files.Get(handle)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid handle"))
	}

	newPos, err := file.Seek(offset, whence)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewNumber(newPos))
}
