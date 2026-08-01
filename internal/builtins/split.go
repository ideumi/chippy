/*
 *
 * Chippy - internal/builtins/split.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"strings"
)

func splitFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("split", 2, "string, delimiter"))
	}

	stringArg, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("split", shared.PositionFirst, shared.TypeString, "string"))
	}

	delimiterArg, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("split", shared.PositionSecond, shared.TypeString, "delimiter"))
	}

	str := stringArg.Value
	delimiter := delimiterArg.Value

	if len(delimiter) == 0 {
		return res.FailAt(2, "Delimiter cannot be empty")
	}

	parts := strings.Split(str, delimiter)

	elements := make([]values.Value, len(parts))

	for i, part := range parts {
		elements[i] = values.NewString(part)
	}

	result := values.NewList(elements)

	return res.Success(result)
}
