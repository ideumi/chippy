/*
 *
 * Chippy - internal/optional/http/urlencode.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"net/url"
)

func urlencodeFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "urlencode"), 1, "text"))
	}

	str, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "urlencode"), shared.TypeString, "text"))
	}

	encoded := url.QueryEscape(str.Value)

	return res.Success(values.NewString(encoded))
}
