/*
 *
 * Chippy - internal/builtins/list.go
 *
 */

package builtins

import (
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"unicode/utf8"
)

func listFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	// Errors are handled differently here, since list() takes an arbitrary amount of args.
	if len(args) == 0 {
		// Create empty list
		return res.Success(values.NewList([]values.Value{}))
	} else if len(args) == 1 {
		// Convert single value to list
		value := args[0]

		if listVal, ok := values.AsList(value); ok {
			elements := make([]values.Value, len(listVal.Elements))
			copy(elements, listVal.Elements)

			return res.Success(values.NewList(elements))
		}

		if str, ok := values.AsString(value); ok {
			elements := make([]values.Value, 0, utf8.RuneCountInString(str.Value))

			for _, char := range str.Value {
				elements = append(elements, values.NewString(string(char)))
			}

			return res.Success(values.NewList(elements))
		}

		if value.IsNumber() {
			return res.Success(values.NewList([]values.Value{value}))
		}

		return res.Success(values.NewString(constants.STR_ERR))
	} else {
		// Create list from multiple arguments
		elements := make([]values.Value, len(args))

		for i, arg := range args {
			elements[i] = arg
		}

		return res.Success(values.NewList(elements))
	}
}
