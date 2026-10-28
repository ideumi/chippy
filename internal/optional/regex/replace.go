/*
 *
 * Chippy - internal/optional/regex/replace.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

func replaceFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 5 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "replace"), 5, "expression, text, replacement, limit, admittance"))
	}

	expression, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "replace"), shared.PositionFirst, shared.TypeString, "expression"))
	}

	text, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "replace"), shared.PositionSecond, shared.TypeString, "text"))
	}

	replacement, ok := values.AsString(args[2])

	if !ok {
		return res.FailAt(3,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "replace"), shared.PositionThird, shared.TypeString, "replacement"))
	}

	ok, admitted, gate := gateExpression(optional.Prefixed(OptionalName, "replace"), args, expression.Value, 3, shared.PositionFourth, 4, shared.PositionFifth)

	if gate.Error != nil {
		return gate
	}

	if !ok {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	re, err := getCompiled(expression.Value, admitted)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(re.ReplaceAllString(text.Value, replacement.Value)))
}
