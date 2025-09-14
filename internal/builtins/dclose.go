/*
 *
 * RR2 - internal/builtins/dclose.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func dcloseFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("dclose", 1, "handle"),
			ctx,
		))
	}

	handleNum, ok := args[0].(*values.Number)
	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("dclose", shared.TypeNumber, "handle"),
			ctx,
		))
	}

	handleID := int(handleNum.Value)
	_, exists := shared.GetDirHandle(handleID)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid directory handle"),
			ctx,
		))
	}

	shared.RemoveDirHandle(handleID)

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
