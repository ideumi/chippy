/*
 *
 * RR2 - internal/builtins/actor.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"fmt"
	"os"
)

func actorFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) < 1 {
		return res.Failure(errors.NewRTError(
			nil, nil,
			shared.Errors.InvalidArgCountWithHint("actor", 1, "function, ...args")))
	}

	fn, ok := args[0].(values.Callable)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("actor", shared.TypeFunction, "handler")))
	}

	fnArgs := args[1:]

	if len(fnArgs) != fn.ArgCount() {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue(fmt.Sprintf("handler '%s' expects %d argument(s), got %d",
				fn.CallableName(), fn.ArgCount(), len(fnArgs)))))
	}

	orch := orchestrator.Get()

	var inheritedOpts []string

	if spawnerInst := orch.GetInstance(ctx.InstanceID); spawnerInst != nil {
		inheritedOpts = orch.GetLoadedOpts(spawnerInst)
	}

	inst := orch.CreateActor()

	fnCopy := fn.Copy()
	argsCopy := make([]values.Value, len(fnArgs))

	for i, arg := range fnArgs {
		argsCopy[i] = arg.Copy()
	}

	globalNames, globalValues := orchestrator.SnapshotUserGlobals(ctx)

	transfer := append([]values.Value{fnCopy}, argsCopy...)
	transfer = append(transfer, globalValues...)

	orchestrator.IsolateForTransfer(transfer...)

	go func() {
		sent := false
		deliver := func(result orchestrator.ActorResult) {
			if sent {
				return
			}

			sent = true
			inst.SendResult(result)
		}

		// A panic in the handler is turned into a result first, then the
		// open handles are closed, and only then is the orchestrator told
		// the actor has finished.

		// Go defer runs in the opposite order to how it is written. So
		// it's all backwards here.
		defer func() {
			orch.MarkFinished(inst)
			orch.CheckDeadlock()
		}()

		defer func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr,
						"chippy: actor %d handle cleanup panicked: %v\n",
						inst.ID, r)
				}
			}()

			inst.Registry.CloseAll()
		}()

		defer func() {
			if r := recover(); r != nil {
				deliver(orchestrator.ActorResult{
					Err: fmt.Errorf("actor panic: %v", r),
				})
			}
		}()

		orch.InitActorModena(inst)

		if inst.Modena == nil {
			deliver(orchestrator.ActorResult{
				Err: fmt.Errorf("failed to create actor runtime"),
			})

			return
		}

		actorCtx := inst.Modena.GetGlobalContext()

		for _, optName := range inheritedOpts {
			if opt, exists := optional.GetOptional(optName); exists {
				InstallOpt(opt, actorCtx)
				orch.AddLoadedOpt(inst, optName)
			}
		}

		for i, name := range globalNames {
			actorCtx.SymbolTable.Set(name, globalValues[i])
		}

		for _, arg := range argsCopy {
			orchestrator.DeepRebindContext(arg, actorCtx)
		}

		for _, val := range globalValues {
			orchestrator.DeepRebindContext(val, actorCtx)
		}

		toBind := append([]values.Value{fnCopy}, argsCopy...)
		toBind = append(toBind, globalValues...)

		orchestrator.BindValuesToGlobals(toBind, actorCtx)

		result := fnCopy.Execute(argsCopy)

		if result.Error != nil {
			deliver(orchestrator.ActorResult{Err: result.Error})

			return
		}

		var returnValue values.Value

		if result.FuncReturnValue != nil {
			returnValue = result.FuncReturnValue
		} else {
			returnValue = result.Value
		}

		// A returned function can still point at variables belonging to
		// the actor, isolate before sending it back.
		if returnValue != nil {
			orchestrator.IsolateForTransfer(returnValue)
		}

		deliver(orchestrator.ActorResult{Value: returnValue})
	}()

	return res.Success(values.NewNumber(inst.ID).SetContext(ctx))
}
