/*
 *
 * RR2 - internal/optional/hash/sha256.go
 *
 */

package hash

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"

	"crypto/sha256"
)

func sha256Function(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "sha256"), 1, "bytes"),
			ctx,
		))
	}

	bytesVal, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "sha256"), shared.TypeBytes, "bytes"),
			ctx,
		))
	}

	hash := sha256.Sum256(bytesVal.Data)

	return res.Success(values.NewBytes(hash[:]))
}
