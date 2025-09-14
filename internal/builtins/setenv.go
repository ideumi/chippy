/*
 *
 * RR2 - internal/builtins/setenv.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
)

func setenvFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("setenv", 2, "name, value"),
			ctx,
		))
	}

	nameStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("setenv", shared.PositionFirst, shared.TypeString, "name"),
			ctx,
		))
	}

	valueStr, ok := args[1].(*values.String)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("setenv", shared.PositionSecond, shared.TypeString, "value"),
			ctx,
		))
	}

	// Set environment variable
	err := os.Setenv(nameStr.Value, valueStr.Value)

	if err != nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Failed to set environment variable: "+err.Error(),
			ctx,
		))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
