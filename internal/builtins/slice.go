/*
 *
 * RR2 - internal/builtins/slice.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func sliceFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("slice", 3, "container, start, end")))
	}

	switch container := args[0].(type) {
	case *values.String:
		startNum, ok := args[1].(*values.Number)

		if !ok {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionSecond, shared.TypeNumber, "start")))
		}

		endNum, ok := args[2].(*values.Number)

		if !ok {
			posStart, posEnd := args[2].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionThird, shared.TypeNumber, "end")))
		}

		start := int(startNum.Value)
		end := int(endNum.Value)

		if start < 1 {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Start must be >= 1"))
		}

		runes := []rune(container.Value)
		size := len(runes)

		if start > size || end < start {
			return res.Success(values.NewString("").SetContext(ctx))
		}

		if end > size {
			end = size
		}

		return res.Success(values.NewString(string(runes[start-1 : end])).SetContext(ctx))

	case *values.Bytes:
		startNum, ok := args[1].(*values.Number)

		if !ok {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionSecond, shared.TypeNumber, "start")))
		}

		endNum, ok := args[2].(*values.Number)

		if !ok {
			posStart, posEnd := args[2].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionThird, shared.TypeNumber, "end")))
		}

		start := int(startNum.Value)
		end := int(endNum.Value)

		if start < 1 {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Start must be >= 1"))
		}

		data := container.Data
		size := len(data)

		if start > size || end < start {
			return res.Success(values.NewBytes([]byte{}).SetContext(ctx))
		}

		if end > size {
			end = size
		}

		result := make([]byte, end-(start-1))
		copy(result, data[start-1:end])

		return res.Success(values.NewBytes(result).SetContext(ctx))

	case *values.List:
		startNum, ok := args[1].(*values.Number)

		if !ok {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionSecond, shared.TypeNumber, "start")))
		}

		endNum, ok := args[2].(*values.Number)

		if !ok {
			posStart, posEnd := args[2].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionThird, shared.TypeNumber, "end")))
		}

		start := int(startNum.Value)
		end := int(endNum.Value)

		if start < 1 {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Start must be >= 1"))
		}

		elements := container.Elements
		size := len(elements)

		if start > size || end < start {
			return res.Success(values.NewList([]values.Value{}).SetContext(ctx))
		}

		if end > size {
			end = size
		}

		result := make([]values.Value, end-(start-1))
		copy(result, elements[start-1:end])

		return res.Success(values.NewList(result).SetContext(ctx))

	default:
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionFirst, "a string, bytes, or list", "container")))
	}
}
