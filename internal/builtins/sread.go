/*
 *
 * RR2 - internal/builtins/sread.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"io"
)

func sreadFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("sread", 2, "handle, maxBytes"),
			ctx,
		))
	}

	// Get handle argument
	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("sread", shared.PositionFirst, shared.TypeNumber, "handle"),
			ctx,
		))
	}

	// Get maxBytes argument
	maxBytesNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("sread", shared.PositionSecond, shared.TypeNumber, "maxBytes"),
			ctx,
		))
	}

	handle := int(handleNum.Value)
	maxBytes := int(maxBytesNum.Value)

	// Validate maxBytes
	if maxBytes <= 0 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("maxBytes must be greater than 0"),
			ctx,
		))
	}

	// Get socket handle
	socket, exists := shared.GetSocketHandle(handle)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid socket handle"),
			ctx,
		))
	}

	// Prepare buffer
	buffer := make([]byte, maxBytes)
	var n int
	var err error

	// Read based on socket type
	switch socket.Mode {

	case "tcp":
		if socket.Conn == nil {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Socket connection is closed"),
				ctx,
			))
		}

		n, err = socket.Conn.Read(buffer)

	case "udp":
		if socket.UdpConn == nil {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("UDP connection is closed"),
				ctx,
			))
		}

		n, err = socket.UdpConn.Read(buffer)

	case "listen":
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Cannot read from listening socket. Use saccept() first"),
			ctx,
		))

	default:
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid socket mode"),
			ctx,
		))
	}

	// Handle read errors
	if err != nil {
		if err == io.EOF {
			// EOF
			return res.Success(values.NewBytes([]byte{}).SetContext(ctx))
		}

		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	if n == 0 {
		return res.Success(values.NewBytes([]byte{}).SetContext(ctx))
	}

	return res.Success(values.NewBytes(buffer[:n]).SetContext(ctx))
}
