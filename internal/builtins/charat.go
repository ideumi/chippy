/*
 *
 * Chippy - internal/builtins/charat.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func charatFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("charat", 2, "string, index"))
	}

	stringArg, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("charat", shared.PositionFirst, shared.TypeString, "string"))
	}

	indexNum := args[1]

	if !indexNum.IsNumber() {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("charat", shared.PositionSecond, shared.TypeNumber, "index"))
	}

	idx64, err := indexNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	char, inRange := stringArg.RuneAt(int(idx64))

	if !inRange {
		return res.FailAt(2, "Index out of bounds")
	}

	return res.Success(char)
}
