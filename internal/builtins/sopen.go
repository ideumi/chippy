/*
 *
 * RR2 - internal/builtins/sopen.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/handles"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"fmt"
	"net"
)

func sopenFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("sopen", 3, "address, port, mode"),
			ctx,
		))
	}

	// Get address argument
	addressStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("sopen", shared.PositionFirst, shared.TypeString, "address"),
			ctx,
		))
	}

	// Get port argument
	portNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("sopen", shared.PositionSecond, shared.TypeNumber, "port"),
			ctx,
		))
	}

	// Get mode argument
	modeStr, ok := args[2].(*values.String)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("sopen", shared.PositionThird, shared.TypeString, "mode"),
			ctx,
		))
	}

	address := addressStr.Value
	port := int(portNum.Value)
	mode := modeStr.Value

	// Validate mode
	if mode != "tcp" && mode != "udp" && mode != "listen" {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid mode. Use 'tcp', 'udp', or 'listen'"),
			ctx,
		))
	}

	// Validate port range
	if port < 1 || port > 65535 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Port must be between 1 and 65535"),
			ctx,
		))
	}

	registry := orchestrator.Get().GetRegistry(ctx)
	socket := &handles.SocketHandle{Mode: mode}

	switch mode {

	case "tcp":
		// TCP client connection
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))

		if err != nil {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		socket.Conn = conn

	case "udp":
		// UDP connection
		raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))

		if err != nil {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx)) // Address resolution failed
		}

		conn, err := net.DialUDP("udp", nil, raddr)

		if err != nil {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		socket.UdpConn = conn
		socket.Address = fmt.Sprintf("%s:%d", address, port)

	case "listen":
		// TCP server (bind & listen)
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))

		if err != nil {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		socket.Listener = listener
	}

	handle := registry.Alloc.Alloc()
	registry.Sockets.Store(handle, socket)

	return res.Success(values.NewNumber(float64(handle)).SetContext(ctx))
}
