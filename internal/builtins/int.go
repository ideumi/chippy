/*
 *
 * RR2 - internal/builtins/int.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func intFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("int", 1, "value")))
	}

	num, ok := args[0].(*values.Number)
	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("int", shared.TypeNumber, "value")))
	}

	intVal, err := num.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	return res.Success(values.NewNumber(intVal).SetContext(ctx))
}
