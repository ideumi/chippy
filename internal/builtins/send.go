/*
 *
 * RR2 - internal/builtins/send.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func sendFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("send", 2, "handle, value"))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("send", shared.PositionFirst, shared.TypeNumber, "handle"))
	}

	id64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	instanceID := int(id64)

	inst := orchestrator.Get().GetInstance(instanceID)

	if inst == nil {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid actor handle"))
	}

	inst.Inbox.Send(args[1].Copy())

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
