/*
 *
 * RR2 - internal/optional/json/stringify.go
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

func stringifyFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "stringify"), 1, "value"))
	}

	goValue, err := marshalValue(args[0], map[values.Value]bool{}, 0)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	jsonBytes, err := json.Marshal(goValue)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(string(jsonBytes)))
}
