/*
 *
 * RR2 - internal/builtins/error.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func errorFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("error", 1, "error message"),
			ctx,
		))
	}

	message, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("error", shared.TypeString, "error message"),
			ctx,
		))
	}

	posStart, posEnd := args[0].GetPos()

	return res.Failure(errors.NewRTError(
		posStart, posEnd,
		message.Value,
		ctx,
	))
}
