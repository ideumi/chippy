/*
 *
 * RR2 - internal/builtins/indexof.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"strings"
)

func indexofFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("indexof", 2, "haystack, needle"),
			ctx,
		))
	}

	haystackArg, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("indexof", shared.PositionFirst, shared.TypeString, "haystack"),
			ctx,
		))
	}

	needleArg, ok := args[1].(*values.String)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("indexof", shared.PositionSecond, shared.TypeString, "needle"),
			ctx,
		))
	}

	haystack := haystackArg.Value
	needle := needleArg.Value

	if len(needle) == 0 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Cannot search for empty string",
			ctx,
		))
	}

	index := strings.Index(haystack, needle)

	if index == -1 {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewNumber(float64(index)).SetContext(ctx))
}
