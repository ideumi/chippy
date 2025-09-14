/*
 *
 * RR2 - internal/builtins/get.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func getFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("get", 2, "list, index"),
			ctx,
		))
	}

	value := args[0]

	indexNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("get", shared.PositionSecond, shared.TypeNumber, "index"),
			ctx,
		))
	}

	index := int(indexNum.Value)

	switch v := value.(type) {
	case *values.Bytes:
		if index < 0 || index >= len(v.Data) {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Index out of bounds",
				ctx,
			))
		}

		return res.Success(values.NewNumber(float64(v.Data[index])).SetContext(ctx))

	case *values.List:
		if index < 0 || index >= len(v.Elements) {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Index out of bounds",
				ctx,
			))
		}

		return res.Success(v.Elements[index])

	default:
		posStart, posEnd := args[0].GetPos()
		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("get", shared.PositionFirst, shared.TypeListOrBytes, shared.TypeListOrBytes),
			ctx,
		))
	}
}
