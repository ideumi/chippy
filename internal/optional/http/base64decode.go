/*
 *
 * RR2 - internal/optional/http/base64decode.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"encoding/base64"
)

func base64decodeFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "base64decode"), 1, "text"),
			ctx,
		))
	}

	str, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "base64decode"), shared.TypeString, "text"),
			ctx,
		))
	}

	decoded, err := base64.StdEncoding.DecodeString(str.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewBytes(decoded).SetContext(ctx))
}
