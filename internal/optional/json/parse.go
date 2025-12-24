/*
 *
 * RR2 - internal/optional/json/parse.go
 *
 */

package json

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"encoding/json"
)

func parseFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "parse"), 1, "jsonStr"),
			ctx,
		))
	}

	jsonStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "parse"), shared.TypeString, "jsonStr"),
			ctx,
		))
	}

	var raw interface{}

	if err := json.Unmarshal([]byte(jsonStr.Value), &raw); err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	result := goToLibMap(raw, ctx)
	return res.Success(result)
}
