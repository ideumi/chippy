/*
 *
 * RR2 - internal/builtins/iserr.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
)

func iserrFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("iserr", 1, "value"))
	}

	value := args[0]

	if str, ok := value.(*values.String); ok {
		if str.Value == constants.STR_ERR {
			return res.Success(values.NewNumber(constants.NUM_TRU).SetContext(ctx))
		}
	}

	return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))
}
