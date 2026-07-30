/*
 *
 * RR2 - internal/builtins/wait.go
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

func waitFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("wait", 1, "handle"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("wait", shared.TypeNumber, "handle"))
	}

	id64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	instanceID := int(id64)

	if instanceID == 0 {
		return res.FailAt(1, shared.Errors.InvalidValue("Cannot wait on the main actor (handle 0)"))
	}

	orch := orchestrator.Get()
	inst := orch.GetInstance(instanceID)

	if inst == nil {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid actor handle"))
	}

	if !orch.MarkWaited(inst) {
		return res.FailAt(1, shared.Errors.InvalidValue("Actor has already been waited on"))
	}

	selfID := ctx.InstanceID
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
		return res.FailAt(1, shared.Errors.InvalidValue("Deadlock: wait() target cannot finish"))
	}

	orch.RemoveInstance(instanceID)

	if result.Err != nil {
		// Actor-body RTErrors carry their own internal position. Pass through.
		// Panics surface as plain errors (see actor.go) and have none, so wrap
		// them with the wait call site so reporting isn't blind.
		if _, ok := result.Err.(*errors.RTError); ok {
			return res.Failure(result.Err)
		}

		return res.FailAt(1, result.Err.Error())
	}

	if result.Value.IsSet() {
		// The actor already isolated returned closures from its own scope
		// (actor.go calls IsolateForTransfer on returnValue before delivering).
		// Rebind them onto the waiter's globals so the caller can actually use
		// them. Do not SetContext on the outer value: for a Function, ctx IS
		// the captured scope, so SetContext would overwrite the isolation we
		// just relied on.
		orchestrator.BindValuesToGlobals(
			[]values.Value{result.Value},
			selfInst.Modena.GetGlobalContext(),
		)

		return res.Success(result.Value)
	}

	return res.Success(values.NewNumber(constants.NUM_NUL))
}
