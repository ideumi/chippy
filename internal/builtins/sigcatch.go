/*
 *
 * RR2 - internal/builtins/sigcatch.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os/signal"
	"syscall"
)

func sigcatchFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sigcatch", 1, "signums"))
	}

	listArg, ok := args[0].(*values.List)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("sigcatch", shared.TypeList, "signums"))
	}

	sigs := make([]syscall.Signal, 0, len(listArg.Elements))

	for _, el := range listArg.Elements {
		num, ok := el.(*values.Number)

		if !ok {
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
		if caught[sig] {
			continue
		}

		caught[sig] = true
		signal.Notify(signalCh, sig)
	}

	signalMu.Unlock()

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
