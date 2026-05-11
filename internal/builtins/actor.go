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

	fn, ok := args[0].(*values.Function)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("actor", shared.PositionFirst, shared.TypeFunction, "handler")))
	}

	fnArgs := args[1:]

	if len(fnArgs) != len(fn.ArgNames) {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue(fmt.Sprintf("handler '%s' expects %d argument(s), got %d",
				fn.Name, len(fn.ArgNames), len(fnArgs)))))
	}

	orch := orchestrator.Get()

	var inheritedOpts []string

	if spawnerInst := orch.GetInstance(ctx.InstanceID); spawnerInst != nil {
		inheritedOpts = orch.GetLoadedOpts(spawnerInst)
	}

	inst := orch.CreateActor()

	fnCopy := fn.Copy().(*values.Function)
	argsCopy := make([]values.Value, len(fnArgs))

	for i, arg := range fnArgs {
		argsCopy[i] = arg.Copy()
	}

	// The handler and its args were copied from the spawner but may still
	// reference the spawner's context. Isolate them so the actor can't reach
	// back into spawner scope. The cycles map is shared so references into
	// the same context get rewritten consistently across all values.
	cycles := make(map[values.Ctx]values.Ctx)
	orchestrator.IsolateForTransfer(fnCopy, cycles)

	for _, arg := range argsCopy {
		orchestrator.IsolateForTransfer(arg, cycles)
	}

	go func() {
		sent := false
		deliver := func(result orchestrator.ActorResult) {
			if sent {
				return
			}

			sent = true
			inst.SendResult(result)
		}

		// Order matters: innermost recover converts a body panic into a
		// result, then handles are closed, then the orchestrator is told
		// we're finished. Declared in reverse so LIFO gives that order.
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

		orch.InitActorRR2(inst)

		if inst.RR == nil {
			deliver(orchestrator.ActorResult{
				Err: fmt.Errorf("failed to create actor runtime"),
			})

			return
		}

		actorCtx := inst.RR.GetGlobalContext()

		for _, optName := range inheritedOpts {
			if opt, exists := optional.GetOptional(optName); exists {
				InstallOpt(opt, actorCtx)
				orch.AddLoadedOpt(inst, optName)
			}
		}

		for _, arg := range argsCopy {
			orchestrator.DeepRebindContext(arg, actorCtx)
		}

		toBind := append([]values.Value{fnCopy}, argsCopy...)
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

		// If the actor returns a closure, it may still reference the
		// actor's context. Isolate it before delivery so the receiver
		// can use it without reaching back into this actor's scope.
		if returnValue != nil {
			orchestrator.IsolateForTransfer(returnValue, make(map[values.Ctx]values.Ctx))
		}

		deliver(orchestrator.ActorResult{Value: returnValue})
	}()

	return res.Success(values.NewNumber(float64(inst.ID)).SetContext(ctx))
}
