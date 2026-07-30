/*
 *
 * RR2 - internal/builtins/num.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"strconv"
)

func numFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("num", 1, "value"))
	}

	value := args[0]

	if value.IsNumber() {
		return res.Success(value.Copy())
	}

	if str, ok := values.AsString(value); ok {
		if floatVal, parseErr := strconv.ParseFloat(str.Value, 64); parseErr == nil {
			num, err := values.NewNumberFromFloat(floatVal)

			if err != nil {
				return res.Success(values.NewString(constants.STR_ERR))
			}

			return res.Success(num)
		}

		return res.Success(values.NewString(constants.STR_ERR))
	}

	if list, ok := values.AsList(value); ok {
		// Trying our best
		if len(list.Elements) == 0 {
			return res.Success(values.NewNumber(constants.NUM_NUL))
		} else if len(list.Elements) == 1 {
			// Try to convert single element
			if elem := list.Elements[0]; elem.IsNumber() {
				return res.Success(elem.Copy())
			}
		}

		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(constants.STR_ERR))
}
