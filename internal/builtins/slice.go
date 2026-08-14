/*
 *
 * Chippy - internal/builtins/slice.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func sliceFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("slice", 3, "container, start, end"))
	}

	str, isText := values.AsString(args[0])
	list, isList := values.AsList(args[0])
	data, isBytes := values.AsBytes(args[0])

	if !isText && !isList && !isBytes {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionFirst, "a string, list, or bytes", "container"))
	}

	start, end, failure, ok := sliceBounds(args)

	if !ok {
		return failure
	}

	if isText {
		last, inRange := clampSliceEnd(start, end, str.RuneCount())

		if !inRange {
			return res.Success(values.NewString(""))
		}

		return res.Success(str.RuneSlice(start, last))
	}

	if isList {
		last, inRange := clampSliceEnd(start, end, len(list.Elements))

		if !inRange {
			return res.Success(values.NewList([]values.Value{}))
		}

		elements := make([]values.Value, last-start+1)
		copy(elements, list.Elements[start-1:last])

		return res.Success(values.NewList(elements))
	}

	last, inRange := clampSliceEnd(start, end, len(data.Data))

	if !inRange {
		return res.Success(values.NewBytes([]byte{}))
	}

	result := make([]byte, last-start+1)
	copy(result, data.Data[start-1:last])

	return res.Success(values.NewBytes(result))
}

func sliceBounds(args []values.Value) (start, end int, failure values.RuntimeResult, ok bool) {
	res := values.NewRuntimeResult()

	if !args[1].IsNumber() {
		return 0, 0, res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionSecond, shared.TypeNumber, "start")), false
	}

	if !args[2].IsNumber() {
		return 0, 0, res.FailAt(3,
			shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionThird, shared.TypeNumber, "end")), false
	}

	start64, err := args[1].AsInt()

	if err != nil {
		return 0, 0, res.Failure(err), false
	}

	end64, err := args[2].AsInt()

	if err != nil {
		return 0, 0, res.Failure(err), false
	}

	if start64 < 1 {
		return 0, 0, res.FailAt(2, "Start must be >= 1"), false
	}

	return int(start64), int(end64), res, true
}

func clampSliceEnd(start, end, size int) (int, bool) {
	if start > size || end < start {
		return 0, false
	}

	if end > size {
		return size, true
	}

	return end, true
}
