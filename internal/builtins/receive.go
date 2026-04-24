/*
 *
 * RR2 - internal/builtins/receive.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func receiveFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("receive", 1, "blocking"),
			ctx,
		))
	}

	blockArg, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("receive", shared.TypeNumber, "blocking"),
			ctx,
		))
	}

	blocking := blockArg.Value != 0

	instanceID := context.GetInstanceID(ctx)
	orch := orchestrator.Get()
	inst := orch.GetInstance(instanceID)

	if inst == nil {
		return res.Failure(errors.NewRTError(
			nil, nil,
			shared.Errors.InvalidValue("Invalid actor handle"),
			ctx,
		))
	}

	globals := inst.RR.GetGlobalContext()

	var items []values.Value

	if blocking {
		var cancelled bool
		items, cancelled = orch.ReceiveBlocking(inst, globals)

		if cancelled {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Deadlock: receive(true) blocked with no possible sender"),
				ctx,
			))
		}
	} else {
		items = inst.Inbox.ReceiveNonBlocking(globals)
	}

	return res.Success(values.NewList(items).SetContext(ctx))
}
