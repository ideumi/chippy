/*
 *
 * RR2 - internal/builtins/transfer.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func transferFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("transfer", 2, "actor, handle"))
	}

	targetNum, ok := args[0].(*values.Number)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("transfer", shared.PositionFirst, shared.TypeNumber, "actor"))
	}

	handleNum, ok := args[1].(*values.Number)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("transfer", shared.PositionSecond, shared.TypeNumber, "handle"))
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
		return res.FailAt(2, shared.Errors.InvalidValue(errMsg))
	}

	return res.Success(values.NewNumber(newID).SetContext(ctx))
}
