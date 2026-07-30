/*
 *
 * RR2 - internal/builtins/keys.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func keysFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("keys", 1, "map"))
	}

	mapVal, ok := values.AsMap(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("keys", shared.TypeMap, "map"))
	}

	elements := make([]values.Value, len(mapVal.Keys))

	for i, key := range mapVal.Keys {
		elements[i] = values.NewString(key)
	}

	return res.Success(values.NewList(elements))
}
