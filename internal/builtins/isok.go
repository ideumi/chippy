/*
 *
 * Chippy - internal/builtins/isok.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
)

func isokFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("isok", 1, "value"))
	}

	value := args[0]

	if str, ok := values.AsString(value); ok {
		if str.Value == constants.STR_OK {
			return res.Success(values.NewNumber(constants.NUM_TRU))
		}
	}

	return res.Success(values.NewNumber(constants.NUM_FAL))
}
