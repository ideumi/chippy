/*
 *
 * RR2 - internal/optional/tls/tlsclose.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

func tlscloseFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "close"), 1, "handle"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(
				optional.Prefixed(OptionalName, "close"), shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)
	tlsHandle, ok := getTLSHandle(ctx, handle)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid TLS handle"))
	}

	if tlsHandle.Closed {
		return res.FailAt(1, shared.Errors.InvalidValue("TLS connection already closed"))
	}

	err = tlsHandle.Conn.Close()

	tlsHandle.Closed = true
	removeTLSHandle(ctx, handle)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(constants.STR_OK))
}
