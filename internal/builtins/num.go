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

func numFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("num", 1, "value"))
	}

	value := args[0]

	switch typed := value.(type) {

	case *values.Number:
		return res.Success(typed.Copy().SetContext(ctx))

	case *values.String:
		if floatVal, parseErr := strconv.ParseFloat(typed.Value, 64); parseErr == nil {
			num, err := values.NewNumberFromFloat(floatVal)

			if err != nil {
				return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
			}

			return res.Success(num.SetContext(ctx))
		}

		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))

	case *values.List:
		// Trying our best
		if len(typed.Elements) == 0 {
			return res.Success(values.NewNumber(constants.NUM_NUL).SetContext(ctx))
		} else if len(typed.Elements) == 1 {
			// Try to convert single element
			if elem := typed.Elements[0]; elem != nil {
				if elemNum, ok := elem.(*values.Number); ok {
					return res.Success(elemNum.Copy().SetContext(ctx))
				}
			}
		}

		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))

	default:
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}
}
