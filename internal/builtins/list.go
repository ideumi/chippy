/*
 *
 * Chippy - internal/builtins/list.go
 *
 */

package builtins

import (
	"chip-go/internal/constants"
	"chip-go/internal/values"
)

func listFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	// Errors are handled differently here, since list() takes an arbitrary
	// amount of args.
	if len(args) == 0 {
		return res.Success(values.NewList([]values.Value{}))
	}

	if len(args) > 1 {
		elements := make([]values.Value, len(args))
		copy(elements, args)

		return res.Success(values.NewList(elements))
	}

	value := args[0]

	if listVal, ok := values.AsList(value); ok {
		elements := make([]values.Value, len(listVal.Elements))
		copy(elements, listVal.Elements)

		return res.Success(values.NewList(elements))
	}

	if str, ok := values.AsString(value); ok {
		return res.Success(values.NewList(str.Runes()))
	}

	if value.IsNumber() {
		return res.Success(values.NewList([]values.Value{value}))
	}

	return res.Success(values.NewString(constants.STR_ERR))
}
