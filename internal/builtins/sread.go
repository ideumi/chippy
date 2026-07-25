/*
 *
 * RR2 - internal/builtins/sread.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"io"
)

func sreadFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sread", 2, "handle, maxBytes"))
	}

	// Get handle argument
	handleNum, ok := args[0].(*values.Number)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("sread", shared.PositionFirst, shared.TypeNumber, "handle"))
	}

	// Get maxBytes argument
	maxBytesNum, ok := args[1].(*values.Number)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("sread", shared.PositionSecond, shared.TypeNumber, "maxBytes"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	maxBytes64, err := maxBytesNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)
	maxBytes := int(maxBytes64)

	// Validate maxBytes
	if maxBytes <= 0 {
		return res.FailAt(2, shared.Errors.InvalidValue("maxBytes must be greater than 0"))
	}

	// Get socket handle
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	socket, exists := registry.Sockets.Get(handle)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid socket handle"))
	}

	// Prepare buffer
	buffer := make([]byte, maxBytes)
	var n int

	// Read based on socket type
	switch socket.Mode {

	case "tcp":
		if socket.Conn == nil {
			return res.FailAt(1, shared.Errors.InvalidValue("Socket connection is closed"))
		}

		n, err = socket.Conn.Read(buffer)

	case "udp":
		if socket.UdpConn == nil {
			return res.FailAt(1, shared.Errors.InvalidValue("UDP connection is closed"))
		}

		n, err = socket.UdpConn.Read(buffer)

	case "listen":
		return res.FailAt(1,
			shared.Errors.InvalidValue("Cannot read from listening socket. Use saccept() first"))

	default:
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid socket mode"))
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
