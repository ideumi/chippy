/*
 *
 * RR2 - internal/builtins/sigfree.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os/signal"
	"syscall"
)

func sigfreeFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("sigfree", 1, "signums")))
	}

	listArg, ok := args[0].(*values.List)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("sigfree", shared.TypeList, "signums")))
	}

	sigs := make([]syscall.Signal, 0, len(listArg.Elements))

	for _, el := range listArg.Elements {
		num, ok := el.(*values.Number)

		if !ok {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("All signums must be numbers")))
		}

		sigs = append(sigs, syscall.Signal(int(num.Value)))
	}

	signalMu.Lock()
	ensureSignalInfra()

	for _, sig := range sigs {
		if !caught[sig] {
			continue
		}

		delete(caught, sig)
		signal.Reset(sig)
	}

	becameEmpty := len(caught) == 0
	signalMu.Unlock()

	// After releasing the last caught signal, any actor blocked in signal()
	// can no longer wake. Re-check deadlock so it cancels cleanly. Must run
	// after signalMu.Unlock() to preserve lock order.
	if becameEmpty {
		orchestrator.Get().CheckDeadlock()
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
