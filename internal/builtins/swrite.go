/*
 *
 * Chippy - internal/builtins/swrite.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func swriteFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("swrite", 2, "data, handle"))
	}

	// Get data argument (bytes)
	bytesVal, ok := values.AsBytes(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("swrite", shared.PositionFirst, shared.TypeBytes, shared.TypeBytes))
	}

	// Get handle argument
	handleNum := args[1]

	if !handleNum.IsNumber() {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("swrite", shared.PositionSecond, shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)

	// Get socket handle
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	socket, exists := registry.Sockets.Get(handle)

	if !exists {
		return res.FailAt(2, shared.Errors.InvalidValue("Invalid socket handle"))
	}

	if len(bytesVal.Data) == 0 {
		return res.Success(values.NewNumber(constants.NUM_NUL))
	}

	// Write based on socket type
	var written int

	switch socket.Mode {
	case "tcp":
		if socket.Conn == nil {
			return res.FailAt(2, shared.Errors.InvalidValue("Socket connection is closed"))
		}
		written, err = socket.Conn.Write(bytesVal.Data)

	case "udp":
		if socket.UdpConn == nil {
			return res.FailAt(2, shared.Errors.InvalidValue("UDP connection is closed"))
		}
		written, err = socket.UdpConn.Write(bytesVal.Data)

	case "listen":
		return res.FailAt(2,
			shared.Errors.InvalidValue("Cannot write to listening socket. Use saccept() first"))

	default:
		return res.FailAt(2, shared.Errors.InvalidValue("Invalid socket mode"))
	}

	// Handle write errors
	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewNumber(written))
}
