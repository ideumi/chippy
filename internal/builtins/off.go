/*
 *
 * RR2 - internal/builtins/off.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"os"
)

func offFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("off", 1, "exitCode"))
	}

	exitCode := args[0]

	if !exitCode.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("off", shared.TypeNumber, "exitCode"))
	}

	code64, err := exitCode.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	code := int(code64)

	if code < 0 || code > 255 {
		return res.FailAt(1, shared.Errors.InvalidValue("Exit code must be between 0 and 255"))
	}

	os.Exit(code)

	return values.RuntimeResult{}
}
