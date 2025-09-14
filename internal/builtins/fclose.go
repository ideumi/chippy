/*
 *
 * RR2 - internal/builtins/fclose.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func fcloseFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("fclose", 1, "handle"),
			ctx,
		))
	}

	handleNum, ok := args[0].(*values.Number)
	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("fclose", shared.TypeNumber, "handle"),
			ctx,
		))
	}

	handle := int(handleNum.Value)

	if handle >= 0 && handle <= 2 {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Cannot close standard handles (0, 1, 2)"),
			ctx,
		))
	}

	file, exists := shared.GetFileHandle(handle)

	if exists {
		shared.RemoveFileHandle(handle)
	}

	if exists {
		// Recycle the handle ID for reuse
		shared.RecycleFileHandle(handle)
	}

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid handle"),
			ctx,
		))
	}

	err := file.Close()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
