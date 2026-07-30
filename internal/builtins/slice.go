/*
 *
 * RR2 - internal/builtins/slice.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"unicode/utf8"
)

func sliceFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("slice", 3, "container, start, end"))
	}

	if container, ok := values.AsString(args[0]); ok {
		startNum := args[1]

		if !startNum.IsNumber() {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionSecond, shared.TypeNumber, "start"))
		}

		endNum := args[2]

		if !endNum.IsNumber() {
			return res.FailAt(3,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionThird, shared.TypeNumber, "end"))
		}

		start64, err := startNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		end64, err := endNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		start := int(start64)
		end := int(end64)

		if start < 1 {
			return res.FailAt(2, "Start must be >= 1")
		}

		size := utf8.RuneCountInString(container.Value)

		if start > size || end < start {
			return res.Success(values.NewString(""))
		}

		if end > size {
			end = size
		}

		return res.Success(values.NewString(container.RuneSlice(start, end)))
	}

	if container, ok := values.AsList(args[0]); ok {
		startNum := args[1]

		if !startNum.IsNumber() {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionSecond, shared.TypeNumber, "start"))
		}

		endNum := args[2]

		if !endNum.IsNumber() {
			return res.FailAt(3,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionThird, shared.TypeNumber, "end"))
		}

		start64, err := startNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		end64, err := endNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		start := int(start64)
		end := int(end64)

		if start < 1 {
			return res.FailAt(2, "Start must be >= 1")
		}

		elements := container.Elements
		size := len(elements)

		if start > size || end < start {
			return res.Success(values.NewList([]values.Value{}))
		}

		if end > size {
			end = size
		}

		result := make([]values.Value, end-(start-1))
		copy(result, elements[start-1:end])

		return res.Success(values.NewList(result))
	}

	if container, ok := values.AsBytes(args[0]); ok {
		startNum := args[1]

		if !startNum.IsNumber() {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionSecond, shared.TypeNumber, "start"))
		}

		endNum := args[2]

		if !endNum.IsNumber() {
			return res.FailAt(3,
				shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionThird, shared.TypeNumber, "end"))
		}

		start64, err := startNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		end64, err := endNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		start := int(start64)
		end := int(end64)

		if start < 1 {
			return res.FailAt(2, "Start must be >= 1")
		}

		data := container.Data
		size := len(data)

		if start > size || end < start {
			return res.Success(values.NewBytes([]byte{}))
		}

		if end > size {
			end = size
		}

		result := make([]byte, end-(start-1))
		copy(result, data[start-1:end])

		return res.Success(values.NewBytes(result))
	}

	return res.FailAt(1,
		shared.Errors.InvalidArgTypePositionalWithHint("slice", shared.PositionFirst, "a string, list, or bytes", "container"))
}
