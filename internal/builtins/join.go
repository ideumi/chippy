/*
 *
 * RR2 - internal/builtins/join.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"strings"
)

func joinFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("join", 2, "array, separator"))
	}

	listArg, ok := args[0].(*values.List)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("join", shared.PositionFirst, shared.TypeList, "array"))
	}

	separatorArg, ok := args[1].(*values.String)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("join", shared.PositionSecond, shared.TypeString, "separator"))
	}

	parts := make([]string, len(listArg.Elements))

	for i, element := range listArg.Elements {
		if strVal, ok := element.(*values.String); ok {
			parts[i] = strVal.Value
		} else {
			parts[i] = element.String()
		}
	}

	result := strings.Join(parts, separatorArg.Value)

	return res.Success(values.NewString(result).SetContext(ctx))
}
