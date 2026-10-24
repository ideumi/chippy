/*
 *
 * Chippy - internal/optional/regex/valid.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"regexp"
)

func validFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "valid"), 1, "expression"))
	}

	expression, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "valid"), shared.PositionFirst, shared.TypeString, "expression"))
	}

	_, err := regexp.Compile(expression.Value)

	if err != nil {
		return res.Success(values.Bool(false))
	}

	return res.Success(values.Bool(true))
}
