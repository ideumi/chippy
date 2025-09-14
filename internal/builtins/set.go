/*
 *
 * RR2 - internal/builtins/set.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func setFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("set", 3, "list, index, value"),
			ctx,
		))
	}

	container := args[0]

	indexNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("set", shared.PositionSecond, shared.TypeNumber, "index"),
			ctx,
		))
	}

	index := int(indexNum.Value)

	switch v := container.(type) {
	case *values.Bytes:
		valueNum, ok := args[2].(*values.Number)
		if !ok {
			posStart, posEnd := args[2].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("set", shared.PositionThird, shared.TypeNumber, "value"),
				ctx,
			))
		}

		if index < 0 || index >= len(v.Data) {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Index out of bounds",
				ctx,
			))
		}

		byteValue := int(valueNum.Value)
		if byteValue < 0 || byteValue > 255 {
			posStart, posEnd := args[2].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Byte values must be between 0 and 255",
				ctx,
			))
		}

		newBytes := v.Copy().(*values.Bytes)
		newBytes.Data[index] = byte(byteValue)

		return res.Success(newBytes.SetContext(ctx))

	case *values.List:
		if index < 0 || index >= len(v.Elements) {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Index out of bounds",
				ctx,
			))
		}

		newList := v.Copy().(*values.List)
		newList.Elements[index] = args[2].SetContext(ctx)

		return res.Success(newList)

	default:
		posStart, posEnd := args[0].GetPos()
		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("set", shared.PositionFirst, shared.TypeListOrBytes, shared.TypeListOrBytes),
			ctx,
		))
	}
}
