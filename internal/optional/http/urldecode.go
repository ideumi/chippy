/*
 *
 * RR2 - internal/optional/http/urldecode.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"net/url"
)

func urldecodeFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "urldecode"), 1, "text"))
	}

	str, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "urldecode"), shared.TypeString, "text"))
	}

	decoded, err := url.QueryUnescape(str.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(decoded).SetContext(ctx))
}
