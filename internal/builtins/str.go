/*
 *
 * RR2 - internal/builtins/str.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func strFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("str", 1, "value"))
	}

	value := args[0]

	if str, ok := values.AsString(value); ok {
		return res.Success(values.NewString(str.Value))
	}

	return res.Success(values.NewString(value.String()))
}
