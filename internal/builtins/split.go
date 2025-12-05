/*
 *
 * RR2 - internal/builtins/split.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"strings"
)

func splitFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("split", 2, "string, delimiter"),
			ctx,
		))
	}

	stringArg, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("split", shared.PositionFirst, shared.TypeString, "string"),
			ctx,
		))
	}

	delimiterArg, ok := args[1].(*values.String)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("split", shared.PositionSecond, shared.TypeString, "delimiter"),
			ctx,
		))
	}

	str := stringArg.Value
	delimiter := delimiterArg.Value

	if len(delimiter) == 0 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Delimiter cannot be empty",
			ctx,
		))
	}

	parts := strings.Split(str, delimiter)

	elements := make([]values.Value, len(parts))

	for i, part := range parts {
		elements[i] = values.NewString(part).SetContext(ctx)
	}

	result := values.NewList(elements)

	return res.Success(result.SetContext(ctx))
}
