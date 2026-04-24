/*
 *
 * RR2 - internal/builtins/wait.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func waitFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("wait", 1, "handle"),
			ctx,
		))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("wait", shared.TypeNumber, "handle"),
			ctx,
		))
	}

	instanceID := int(handleNum.Value)

	if instanceID == 0 {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Cannot wait on the main actor (handle 0)"),
			ctx,
		))
	}

	orch := orchestrator.Get()
	inst := orch.GetInstance(instanceID)

	if inst == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid actor handle"),
			ctx,
		))
	}

	if !orch.MarkWaited(inst) {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Actor has already been waited on"),
			ctx,
		))
	}

	selfID := context.GetInstanceID(ctx)
	selfInst := orch.GetInstance(selfID)

	cancelCh := orch.BeginBlocking(selfInst, orchestrator.StateBlockedWait)
	orch.CheckDeadlock()

	var result orchestrator.ActorResult
	gotResult := false

	select {
	case result = <-inst.ResultCh:
		gotResult = true
	case <-cancelCh:
		// Cancelled by deadlock detector. Prefer a pending result over
		// reporting a deadlock, actor may have delivered between the
		// cancel close and our select firing.
		select {
		case result = <-inst.ResultCh:
			gotResult = true
		default:
		}
	}

	orch.EndBlocking(selfInst)

	if !gotResult {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Deadlock: wait() target cannot finish"),
			ctx,
		))
	}

	orch.RemoveInstance(instanceID)

	if result.Err != nil {
		// Actor-body RTErrors carry their own internal position; pass through.
		// Panics surface as plain errors (see actor.go) and have none, so wrap
		// them with the wait call site so reporting isn't blind.
		if _, ok := result.Err.(*errors.RTError); ok {
			return res.Failure(result.Err)
		}

		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			result.Err.Error(),
			ctx,
		))
	}

	if result.Value != nil {
		// The actor already isolated returned closures from its own scope
		// (actor.go calls IsolateForTransfer on returnValue before delivering).
		// Rebind them onto the waiter's globals so the caller can actually use
		// them. Do not SetContext on the outer value: for a Function, ctx IS
		// the captured scope, so SetContext would overwrite the isolation we
		// just relied on.
		orchestrator.BindValuesToGlobals(
			[]values.Value{result.Value},
			selfInst.RR.GetGlobalContext(),
		)

		return res.Success(result.Value)
	}

	return res.Success(values.NewNumber(constants.NUM_NUL).SetContext(ctx))
}
