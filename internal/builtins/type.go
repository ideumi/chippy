/*
 *
 * RR2 - internal/builtins/type.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func typeFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("type", 1, "value"))
	}

	value := args[0]
	typeName := "unknown"

	if value.IsNumber() {
		typeName = "number"
	} else if _, ok := values.AsString(value); ok {
		typeName = "string"
	} else if _, ok := values.AsList(value); ok {
		typeName = "list"
	} else if _, ok := values.AsBytes(value); ok {
		typeName = "bytes"
	} else if _, ok := values.AsMap(value); ok {
		typeName = "map"
	} else if _, ok := values.AsBuiltIn(value); ok {
		typeName = "builtin"
	} else if _, ok := values.AsCallable(value); ok {
		typeName = "function"
	}

	return res.Success(values.NewString(typeName))
}
