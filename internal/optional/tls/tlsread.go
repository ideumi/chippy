/*
 *
 * Chippy - internal/optional/tls/tlsread.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"io"
)

func tlsreadFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "read"), 2, "handle, maxBytes"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "read"), shared.PositionFirst, shared.TypeNumber, "handle"))
	}

	maxBytesNum := args[1]

	if !maxBytesNum.IsNumber() {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "read"), shared.PositionSecond, shared.TypeNumber, "maxBytes"))
	}

	max64, err := maxBytesNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	maxBytes := int(max64)

	if maxBytes <= 0 {
		return res.FailAt(2, shared.Errors.InvalidValue("maxBytes must be greater than 0"))
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
		return res.FailAt(1, shared.Errors.InvalidValue("TLS connection is closed"))
	}

	buffer := make([]byte, maxBytes)
	n, err := tlsHandle.Conn.Read(buffer)

	// If we got data, return it (even if EOF came with it)
	if n > 0 {
		return res.Success(values.NewBytes(buffer[:n]))
	}

	// n == 0: no data read
	// Treat EOF and explicit "no error" as clean connection close
	if err == nil || err == io.EOF {
		return res.Success(values.NewBytes([]byte{}))
	}

	// Any other error is a problem
	return res.Success(values.NewString(constants.STR_ERR))
}
