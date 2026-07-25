/*
 *
 * RR2 - internal/builtins/charat.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func charatFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("charat", 2, "string, index"))
	}

	stringArg, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("charat", shared.PositionFirst, shared.TypeString, "string"))
	}

	indexNum, ok := args[1].(*values.Number)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("charat", shared.PositionSecond, shared.TypeNumber, "index"))
	}

	idx64, err := indexNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	index := int(idx64)
	str := stringArg.Value

	// UTF8
	runes := []rune(str)

	if index < 1 || index > len(runes) {
		return res.FailAt(2, "Index out of bounds")
	}

	char := string(runes[index-1])

	return res.Success(values.NewString(char).SetContext(ctx))
}
