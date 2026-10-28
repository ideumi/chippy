/*
 *
 * Chippy - internal/optional/regex/groupindex.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

func groupIndexFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 4 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "groupindex"), 4, "expression, name, limit, admittance"))
	}

	expression, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "groupindex"), shared.PositionFirst, shared.TypeString, "expression"))
	}

	name, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "groupindex"), shared.PositionSecond, shared.TypeString, "name"))
	}

	ok, admitted, gate := gateExpression(optional.Prefixed(OptionalName, "groupindex"), args, expression.Value, 2, shared.PositionThird, 3, shared.PositionFourth)

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

	for i := range names {
		if groupKey(names, i) == name.Value {
			return res.Success(values.NewNumber(i + 1))
		}
	}

	return res.Success(values.NewString(constants.STR_ERR))
}
