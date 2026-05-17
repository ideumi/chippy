/*
 *
 * RR2 - internal/builtins/replace.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"strings"
)

func replaceFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("replace", 3, "haystack, needle, replacement")))
	}

	haystackArg, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("replace", shared.PositionFirst, shared.TypeString, "haystack")))
	}

	needleArg, ok := args[1].(*values.String)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("replace", shared.PositionSecond, shared.TypeString, "needle")))
	}

	replacementArg, ok := args[2].(*values.String)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("replace", shared.PositionThird, shared.TypeString, "replacement")))
	}

	haystack := haystackArg.Value
	needle := needleArg.Value
	replacement := replacementArg.Value

	if len(needle) == 0 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Cannot replace empty string"))
	}

	result := strings.ReplaceAll(haystack, needle, replacement)

	return res.Success(values.NewString(result).SetContext(ctx))
}
