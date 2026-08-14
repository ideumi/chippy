/*
 *
 * Chippy - internal/builtins/sigfree.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os/signal"
	"syscall"
)

func sigfreeFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sigfree", 1, "signums"))
	}

	listArg, ok := values.AsList(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("sigfree", shared.TypeList, "signums"))
	}

	sigs := make([]syscall.Signal, 0, len(listArg.Elements))

	for _, el := range listArg.Elements {
		num := el

		if !num.IsNumber() {
			return res.FailAt(1, shared.Errors.InvalidValue("All signums must be numbers"))
		}

		sig64, err := num.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		sigs = append(sigs, syscall.Signal(int(sig64)))
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
		orchestrator.Get().Recheck()
	}

	return res.Success(values.NewString(constants.STR_OK))
}
