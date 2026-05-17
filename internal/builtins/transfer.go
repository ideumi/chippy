/*
 *
 * RR2 - internal/builtins/transfer.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func transferFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("transfer", 2, "actor, handle")))
	}

	targetNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("transfer", shared.PositionFirst, shared.TypeNumber, "actor")))
	}

	handleNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("transfer", shared.PositionSecond, shared.TypeNumber, "handle")))
	}

	target64, err := targetNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	targetID := int(target64)
	handleID := int(handle64)

	newID, errMsg := orchestrator.Get().Transfer(ctx.InstanceID, targetID, handleID)

	if errMsg != "" {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue(errMsg)))
	}

	return res.Success(values.NewNumber(newID).SetContext(ctx))
}
