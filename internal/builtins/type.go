/*
 *
 * RR2 - internal/builtins/type.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func typeFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("type", 1, "value"),
			ctx,
		))
	}

	value := args[0]
	var typeName string

	switch value.(type) {
	case *values.Number:
		typeName = "number"

	case *values.String:
		typeName = "string"

	case *values.List:
		typeName = "list"

	case *values.Bytes:
		typeName = "bytes"

	case *values.Function:
		typeName = "function"

	case *values.NativeFunction:
		nf := value.(*values.NativeFunction)
		if nf.Plugin {
			typeName = "plugin"
		} else {
			typeName = "builtin"
		}

	default:
		typeName = "unknown"
	}

	return res.Success(values.NewString(typeName).SetContext(ctx))
}
