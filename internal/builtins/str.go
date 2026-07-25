/*
 *
 * RR2 - internal/builtins/str.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func strFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("str", 1, "value"))
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
