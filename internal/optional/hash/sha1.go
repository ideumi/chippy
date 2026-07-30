/*
 *
 * RR2 - internal/optional/hash/sha1.go
 *
 */

package hash

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"

	"crypto/sha1"
)

func sha1Function(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "sha1"), 1, "bytes"))
	}

	bytesVal, ok := values.AsBytes(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "sha1"), shared.TypeBytes, "bytes"))
	}

	hash := sha1.Sum(bytesVal.Data)

	return res.Success(values.NewBytes(hash[:]))
}
