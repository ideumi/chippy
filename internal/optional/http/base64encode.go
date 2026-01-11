/*
 *
 * RR2 - internal/optional/http/base64encode.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"encoding/base64"
)

func base64encodeFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "base64encode"), 1, "bytes"),
			ctx,
		))
	}

	dataBytes, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "base64encode"), shared.TypeBytes, "bytes"),
			ctx,
		))
	}

	encoded := base64.StdEncoding.EncodeToString(dataBytes.Data)

	return res.Success(values.NewString(encoded).SetContext(ctx))
}
