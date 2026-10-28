/*
 *
 * Chippy - internal/optional/regex/findindex.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

func findIndexFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 4 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "findindex"), 4, "expression, text, limit, admittance"))
	}

	expression, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "findindex"), shared.PositionFirst, shared.TypeString, "expression"))
	}

	text, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "findindex"), shared.PositionSecond, shared.TypeString, "text"))
	}

	ok, admitted, gate := gateExpression(optional.Prefixed(OptionalName, "findindex"), args, expression.Value, 2, shared.PositionThird, 3, shared.PositionFourth)

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

	span := re.FindStringIndex(text.Value)

	if span == nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	start, end := byteToRunePosition(text.Value, span[0], span[1])

	return res.Success(values.NewMapFromEntries(
		[]string{"start", "end"},
		map[string]values.Value{
			"start": values.NewNumber(start),
			"end":   values.NewNumber(end),
		}))
}
