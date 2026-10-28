/*
 *
 * Chippy - internal/optional/regex/groupnames.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

func groupNamesFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "groupnames"), 3, "expression, limit, admittance"))
	}

	expression, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "groupnames"), shared.PositionFirst, shared.TypeString, "expression"))
	}

	ok, admitted, gate := gateExpression(optional.Prefixed(OptionalName, "groupnames"), args, expression.Value, 1, shared.PositionSecond, 2, shared.PositionThird)

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

	names := re.SubexpNames()

	elements := make([]values.Value, 0, len(names))

	for i := range names {
		elements = append(elements, values.NewString(groupKey(names, i)))
	}

	return res.Success(values.NewList(elements))
}
