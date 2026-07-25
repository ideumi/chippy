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

	switch typed := value.(type) {

	case *values.Number:
		resultStr = typed.String()

	case *values.String:
		resultStr = typed.Value // Remove quotes for str() conversion

	case *values.List:
		resultStr = typed.String()

	case *values.Bytes:
		resultStr = typed.String()

	case *values.Map:
		resultStr = typed.String()

	default:
		resultStr = value.String()
	}

	return res.Success(values.NewString(resultStr).SetContext(ctx))
}
