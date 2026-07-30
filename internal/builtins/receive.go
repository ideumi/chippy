/*
 *
 * RR2 - internal/builtins/receive.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func receiveFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("receive", 1, "blocking"))
	}

	blockArg := args[0]

	if !blockArg.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("receive", shared.TypeNumber, "blocking"))
	}

	blocking := blockArg.IsTrue()

	instanceID := ctx.InstanceID
	orch := orchestrator.Get()
	inst := orch.GetInstance(instanceID)

	if inst == nil {
		return res.Fail(shared.Errors.InvalidValue("Invalid actor handle"))
	}

	globals := inst.Modena.GetGlobalContext()

	var items []values.Value

	if blocking {
		var cancelled bool
		items, cancelled = orch.ReceiveBlocking(inst, globals)

		if cancelled {
			return res.FailAt(1,
				shared.Errors.InvalidValue("Deadlock: receive(true) blocked with no possible sender"))
		}
	} else {
		items = inst.Inbox.ReceiveNonBlocking(globals)
	}

	return res.Success(values.NewList(items))
}
