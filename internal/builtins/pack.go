/*
 *
 * RR2 - internal/builtins/pack.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func packFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("pack", 1, "value"),
			ctx,
		))
	}

	value := args[0]

	switch v := value.(type) {
	case *values.String:
		bytes := []byte(v.Value)

		return res.Success(values.NewBytes(bytes).SetContext(ctx))

	case *values.List:
		bytes := make([]byte, len(v.Elements))

		for i, element := range v.Elements {
			if element == nil {
				bytes[i] = 0
				continue
			}

			num, ok := element.(*values.Number)

			if !ok {
				posStart, posEnd := args[0].GetPos()
				return res.Failure(errors.NewRTError(
					posStart, posEnd,
					"List elements must be numbers representing bytes",
					ctx,
				))
			}

			byteVal := int(num.Value)

			if byteVal < 0 || byteVal > 255 {
				posStart, posEnd := args[0].GetPos()
				return res.Failure(errors.NewRTError(
					posStart, posEnd,
					"Byte values must be between 0 and 255",
					ctx,
				))
			}

			bytes[i] = byte(byteVal)
		}

		return res.Success(values.NewBytes(bytes).SetContext(ctx))

	default:
		posStart, posEnd := args[0].GetPos()
		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("pack", shared.TypeStringOrList, shared.TypeStringOrList),
			ctx,
		))
	}
}
