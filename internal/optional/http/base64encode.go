/*
 *
 * Chippy - internal/optional/http/base64encode.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"encoding/base64"
)

func base64encodeFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "base64encode"), 1, "bytes"))
	}

	dataBytes, ok := values.AsBytes(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(optional.Prefixed(OptionalName, "base64encode"), shared.TypeBytes, "bytes"))
	}

	encoded := base64.StdEncoding.EncodeToString(dataBytes.Data)

	return res.Success(values.NewString(encoded))
}
