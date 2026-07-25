/*
 *
 * RR2 - internal/builtins/sclose.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func scloseFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sclose", 1, "handle"))
	}

	// Get handle argument
	handleNum, ok := args[0].(*values.Number)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("sclose", shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	socket, exists := registry.Sockets.Extract(handle)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid socket handle"))
	}

	registry.Alloc.Free(handle)

	switch socket.Mode {
	case "tcp":
		if socket.Conn != nil {
			err = socket.Conn.Close()
			socket.Conn = nil
		}

	case "udp":
		if socket.UdpConn != nil {
			err = socket.UdpConn.Close()
			socket.UdpConn = nil
		}

	case "listen":
		if socket.Listener != nil {
			err = socket.Listener.Close()
			socket.Listener = nil
		}

	default:
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid socket mode"))
	}

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
