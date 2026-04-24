/*
 *
 * RR2 - internal/builtins/swrite.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func swriteFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("swrite", 2, "data, handle"),
			ctx,
		))
	}

	// Get data argument (bytes)
	bytesVal, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("swrite", shared.PositionFirst, shared.TypeBytes, shared.TypeBytes),
			ctx,
		))
	}

	// Get handle argument
	handleNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("swrite", shared.PositionSecond, shared.TypeNumber, "handle"),
			ctx,
		))
	}

	handle := int(handleNum.Value)

	// Get socket handle
	registry := orchestrator.Get().GetRegistry(ctx)
	socket, exists := registry.Sockets.Get(handle)

	if !exists {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid socket handle"),
			ctx,
		))
	}

	if len(bytesVal.Data) == 0 {
		return res.Success(values.NewNumber(constants.NUM_NUL).SetContext(ctx))
	}

	// Write based on socket type
	var n int
	var err error

	switch socket.Mode {
	case "tcp":
		if socket.Conn == nil {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Socket connection is closed"),
				ctx,
			))
		}
		n, err = socket.Conn.Write(bytesVal.Data)

	case "udp":
		if socket.UdpConn == nil {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("UDP connection is closed"),
				ctx,
			))
		}
		n, err = socket.UdpConn.Write(bytesVal.Data)

	case "listen":
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Cannot write to listening socket. Use saccept() first"),
			ctx,
		))

	default:
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid socket mode"),
			ctx,
		))
	}

	// Handle write errors
	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewNumber(float64(n)).SetContext(ctx))
}
