/*
 *
 * Chippy - internal/optional/regex/findallgroups.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

func findAllGroupsFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 4 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "findallgroups"), 4, "expression, text, limit, admittance"))
	}

	expression, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "findallgroups"), shared.PositionFirst, shared.TypeString, "expression"))
	}

	text, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "findallgroups"), shared.PositionSecond, shared.TypeString, "text"))
	}

	ok, admitted, gate := gateExpression(optional.Prefixed(OptionalName, "findallgroups"), args, expression.Value, 2, shared.PositionThird, 3, shared.PositionFourth)

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

	matches := re.FindAllStringSubmatch(text.Value, -1)

	elements := make([]values.Value, 0, len(matches))

	for _, match := range matches {
		elements = append(elements, values.NewList(groupPairs(re, match)))
	}

	return res.Success(values.NewList(elements))
}
