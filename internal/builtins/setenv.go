/*
 *
 * RR2 - internal/builtins/setenv.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
)

func setenvFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("setenv", 2, "name, value"))
	}

	nameStr, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("setenv", shared.PositionFirst, shared.TypeString, "name"))
	}

	valueStr, ok := args[1].(*values.String)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("setenv", shared.PositionSecond, shared.TypeString, "value"))
	}

	// Set environment variable
	err := os.Setenv(nameStr.Value, valueStr.Value)

	if err != nil {
		return res.FailAt(1, "Failed to set environment variable: "+err.Error())
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
