/*
 *
 * RR2 - internal/optional/http/urlencode.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"net/url"
)

func urlencodeFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "urlencode"), 1, "text"))
	}

	str, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "urlencode"), shared.TypeString, "text"))
	}

	encoded := url.QueryEscape(str.Value)

	return res.Success(values.NewString(encoded).SetContext(ctx))
}
