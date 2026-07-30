/*
 *
 * RR2 - internal/builtins/unpack.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func unpackFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("unpack", 1, "bytes"))
	}

	bytesVal, ok := values.AsBytes(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("unpack", shared.TypeBytes, shared.TypeBytes))
	}

	// Convert bytes to string without UTF-8 validation
	return res.Success(values.NewString(string(bytesVal.Data)))
}
