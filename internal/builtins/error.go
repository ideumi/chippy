/*
 *
 * RR2 - internal/builtins/error.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func errorFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("error", 1, "error message"))
	}

	message, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("error", shared.TypeString, "error message"))
	}

	return res.FailAt(1, message.Value)
}
