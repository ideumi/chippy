/*
 *
 * RR2 - internal/builtins/sclose.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func scloseFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("sclose", 1, "handle"),
			ctx,
		))
	}

	// Get handle argument
	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("sclose", shared.PositionFirst, shared.TypeNumber, "handle"),
			ctx,
		))
	}

	handle := int(handleNum.Value)

	socket, exists := shared.GetSocketHandle(handle)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid socket handle"),
			ctx,
		))
	}

	// Close based on socket type
	var err error

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
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid socket mode"),
			ctx,
		))
	}

	// Remove from handle map and recycle
	shared.RemoveSocketHandle(handle)
	shared.RecycleFileHandle(handle)

	// Return result
	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
