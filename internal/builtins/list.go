/*
 *
 * RR2 - internal/builtins/list.go
 *
 */

package builtins

import (
	"chip-go/internal/constants"
	"chip-go/internal/values"
)

func listFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	// Errors are handled differently here, since list() takes an arbitrary amount of args.
	if len(args) == 0 {
		// Create empty list
		return res.Success(values.NewList([]values.Value{}).SetContext(ctx))
	} else if len(args) == 1 {
		// Convert single value to list
		value := args[0]

		switch v := value.(type) {

		case *values.List:
			return res.Success(v.ShallowCopy().SetContext(ctx))

		case *values.String:
			// Convert string to runes
			runes := []rune(v.Value)
			elements := make([]values.Value, len(runes))

			for i, r := range runes {
				elements[i] = values.NewString(string(r)).SetContext(ctx)
			}

			return res.Success(values.NewList(elements).SetContext(ctx))

		case *values.Number:
			// Create single-element list
			return res.Success(values.NewList([]values.Value{v.SetContext(ctx)}).SetContext(ctx))

		default:
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}
	} else {
		// Create list from multiple arguments
		elements := make([]values.Value, len(args))

		for i, arg := range args {
			elements[i] = arg.SetContext(ctx)
		}

		return res.Success(values.NewList(elements).SetContext(ctx))
	}
}
