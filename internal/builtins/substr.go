/*
 *
 * RR2 - internal/builtins/substr.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func substrFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("substr", 3, "string, start, length"),
			ctx,
		))
	}

	stringArg, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("substr", shared.PositionFirst, shared.TypeString, "string"),
			ctx,
		))
	}

	startNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("substr", shared.PositionSecond, shared.TypeNumber, "start"),
			ctx,
		))
	}

	lengthNum, ok := args[2].(*values.Number)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("substr", shared.PositionThird, shared.TypeNumber, "length"),
			ctx,
		))
	}

	str := stringArg.Value
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

	runes := []rune(str)

	if start > len(runes) {
		return res.Success(values.NewString("").SetContext(ctx))
	}

	end := start - 1 + length

	if end > len(runes) {
		end = len(runes)
	}

	result := string(runes[start-1 : end])

	return res.Success(values.NewString(result).SetContext(ctx))
}
