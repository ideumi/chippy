/*
 *
 * RR2 - internal/builtins/unpack.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func unpackFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("unpack", 1, "bytes"),
			ctx,
		))
	}

	bytesVal, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("unpack", shared.TypeBytes, shared.TypeBytes),
			ctx,
		))
	}

	// Convert bytes to string without UTF-8 validation
	return res.Success(values.NewString(string(bytesVal.Data)).SetContext(ctx))
}
