/*
 *
 * RR2 - internal/builtins/keys.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func keysFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("keys", 1, "map")))
	}

	mapVal, ok := args[0].(*values.Map)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("keys", shared.TypeMap, "map")))
	}

	elements := make([]values.Value, len(mapVal.Keys))

	for i, key := range mapVal.Keys {
		elements[i] = values.NewString(key).SetContext(ctx)
	}

	return res.Success(values.NewList(elements).SetContext(ctx))
}
