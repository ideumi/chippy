/*
 *
 * Chippy - internal/optional/regex/escape.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"regexp"
)

func escapeFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "escape"), 1, "text"))
	}

	text, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "escape"), shared.PositionFirst, shared.TypeString, "text"))
	}

	return res.Success(values.NewString(regexp.QuoteMeta(text.Value)))
}
