/*
 *
 * RR2 - internal/optional/json/parse.go
 *
 */

package json

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"encoding/json"
)

func parseFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "parse"), 1, "jsonStr"))
	}

	jsonStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "parse"), shared.TypeString, "jsonStr"))
	}

	var raw interface{}

	if err := json.Unmarshal([]byte(jsonStr.Value), &raw); err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	result := unmarshalValue(raw, ctx)
	return res.Success(result)
}
