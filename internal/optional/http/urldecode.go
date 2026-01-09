/*
 *
 * RR2 - internal/optional/http/urldecode.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"net/url"
)

func urldecodeFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "urldecode"), 1, "text"),
			ctx,
		))
	}

	str, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "urldecode"), shared.TypeString, "text"),
			ctx,
		))
	}

	decoded, err := url.QueryUnescape(str.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(decoded).SetContext(ctx))
}
