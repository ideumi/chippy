/*
 *
 * RR2 - internal/optional/tls/tlswrite.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

func tlswriteFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "write"), 2, "data, handle"))
	}

	dataBytes, ok := args[0].(*values.Bytes)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "write"), shared.PositionFirst, shared.TypeBytes, "data"))
	}

	handleNum, ok := args[1].(*values.Number)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "write"), shared.PositionSecond, shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)
	tlsHandle, ok := getTLSHandle(ctx, handle)

	if !ok {
		return res.FailAt(2, shared.Errors.InvalidValue("Invalid TLS handle"))
	}

	if tlsHandle.Closed {
		return res.FailAt(2, shared.Errors.InvalidValue("TLS connection is closed"))
	}

	n, err := tlsHandle.Conn.Write(dataBytes.Data)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewNumber(n).SetContext(ctx))
}
