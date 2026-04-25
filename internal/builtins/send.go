/*
 *
 * RR2 - internal/builtins/send.go
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

func sendFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("send", 2, "handle, value"),
			ctx,
		))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("send", shared.PositionFirst, shared.TypeNumber, "handle"),
			ctx,
		))
	}

	instanceID := int(handleNum.Value)

	inst := orchestrator.Get().GetInstance(instanceID)

	if inst == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid actor handle"),
			ctx,
		))
	}

	inst.Inbox.Send(args[1].Copy())

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
