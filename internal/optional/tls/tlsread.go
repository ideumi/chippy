/*
 *
 * RR2 - internal/optional/tls/tlsread.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"io"
)

func tlsreadFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "read"), 2, "handle, maxBytes"),
			ctx,
		))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "read"), shared.PositionFirst, shared.TypeNumber, "handle"),
			ctx,
		))
	}

	maxBytesNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "read"), shared.PositionSecond, shared.TypeNumber, "maxBytes"),
			ctx,
		))
	}

	maxBytes := int(maxBytesNum.Value)

	if maxBytes <= 0 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("maxBytes must be greater than 0"),
			ctx,
		))
	}

	handle := int(handleNum.Value)
	tlsHandle, ok := getTLSHandle(ctx, handle)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid TLS handle"),
			ctx,
		))
	}

	if tlsHandle.Closed {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("TLS connection is closed"),
			ctx,
		))
	}

	buffer := make([]byte, maxBytes)
	n, err := tlsHandle.Conn.Read(buffer)

	// If we got data, return it (even if EOF came with it)
	if n > 0 {
		return res.Success(values.NewBytes(buffer[:n]).SetContext(ctx))
	}

	// n == 0: no data read
	// Treat EOF and explicit "no error" as clean connection close
	if err == nil || err == io.EOF {
		return res.Success(values.NewBytes([]byte{}).SetContext(ctx))
	}

	// Any other error is a problem
	return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
}
