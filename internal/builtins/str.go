/*
 *
 * RR2 - internal/builtins/str.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func strFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("str", 1, "value"),
			ctx,
		))
	}

	value := args[0]
	var resultStr string

	switch v := value.(type) {

	case *values.Number:
		resultStr = v.String()

	case *values.String:
		resultStr = v.Value // Remove quotes for str() conversion

	case *values.List:
		resultStr = v.String()

	case *values.Bytes:
		resultStr = v.String()

	case *values.Map:
		resultStr = v.String()

	default:
		resultStr = value.String()
	}

	return res.Success(values.NewString(resultStr).SetContext(ctx))
}
