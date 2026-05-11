/*
 *
 * RR2 - internal/optional/hash/sha1.go
 *
 */

package hash

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"

	"crypto/sha1"
)

func sha1Function(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "sha1"), 1, "bytes")))
	}

	bytesVal, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "sha1"), shared.TypeBytes, "bytes")))
	}

	hash := sha1.Sum(bytesVal.Data)

	return res.Success(values.NewBytes(hash[:]))
}
