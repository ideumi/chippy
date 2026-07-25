/*
 *
 * RR2 - internal/optional/hash/sha224.go
 *
 */

package hash

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"

	"crypto/sha256"
)

func sha224Function(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "sha224"), 1, "bytes"))
	}

	bytesVal, ok := args[0].(*values.Bytes)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "sha224"), shared.TypeBytes, "bytes"))
	}

	hash := sha256.Sum224(bytesVal.Data)

	return res.Success(values.NewBytes(hash[:]))
}
