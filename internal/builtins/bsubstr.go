/*
 *
 * RR2 - internal/builtins/bsubstr.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func bsubstrFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("bsubstr", 3, "bytes, start, length"),
			ctx,
		))
	}

	bytesArg, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("bsubstr", shared.PositionFirst, shared.TypeBytes, "bytes"),
			ctx,
		))
	}

	startNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("bsubstr", shared.PositionSecond, shared.TypeNumber, "start"),
			ctx,
		))
	}

	lengthNum, ok := args[2].(*values.Number)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("bsubstr", shared.PositionThird, shared.TypeNumber, "length"),
			ctx,
		))
	}

	data := bytesArg.Data
	start := int(startNum.Value)
	length := int(lengthNum.Value)

	if start < 1 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Start must be >= 1",
			ctx,
		))
	}

	if length < 0 {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Length must be non-negative",
			ctx,
		))
	}

	if start > len(data) {
		return res.Success(values.NewBytes([]byte{}).SetContext(ctx))
	}

	end := start - 1 + length

	if end > len(data) {
		end = len(data)
	}

	result := make([]byte, end-(start-1))
	copy(result, data[start-1:end])

	return res.Success(values.NewBytes(result).SetContext(ctx))
}
