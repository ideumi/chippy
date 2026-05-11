/*
 *
 * RR2 - internal/builtins/charat.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func charatFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("charat", 2, "string, index")))
	}

	stringArg, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("charat", shared.PositionFirst, shared.TypeString, "string")))
	}

	indexNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("charat", shared.PositionSecond, shared.TypeNumber, "index")))
	}

	index := int(indexNum.Value)
	str := stringArg.Value

	// UTF8
	runes := []rune(str)

	if index < 1 || index > len(runes) {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Index out of bounds"))
	}

	char := string(runes[index-1])

	return res.Success(values.NewString(char).SetContext(ctx))
}
