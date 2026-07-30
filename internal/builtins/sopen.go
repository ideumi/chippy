/*
 *
 * Chippy - internal/builtins/sopen.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/handles"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"net"
	"strconv"
)

func sopenFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sopen", 3, "address, port, mode"))
	}

	// Get address argument
	addressStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("sopen", shared.PositionFirst, shared.TypeString, "address"))
	}

	// Get port argument
	portNum := args[1]

	if !portNum.IsNumber() {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("sopen", shared.PositionSecond, shared.TypeNumber, "port"))
	}

	// Get mode argument
	modeStr, ok := values.AsString(args[2])

	if !ok {
		return res.FailAt(3,
			shared.Errors.InvalidArgTypePositionalWithHint("sopen", shared.PositionThird, shared.TypeString, "mode"))
	}

	port64, err := portNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	address := addressStr.Value
	port := int(port64)
	mode := modeStr.Value

	// Validate mode
	if mode != "tcp" && mode != "udp" && mode != "listen" {
		return res.FailAt(3, shared.Errors.InvalidValue("Invalid mode. Use 'tcp', 'udp', or 'listen'"))
	}

	// Validate port range
	if port < 1 || port > 65535 {
		return res.FailAt(2, shared.Errors.InvalidValue("Port must be between 1 and 65535"))
	}

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	socket := &handles.SocketHandle{Mode: mode}

	switch mode {

	case "tcp":
		// TCP client connection
		conn, err := net.Dial("tcp", net.JoinHostPort(address, strconv.Itoa(port)))

		if err != nil {
			return res.Success(values.NewString(constants.STR_ERR))
		}

		socket.Conn = conn

	case "udp":
		// UDP connection
		raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(address, strconv.Itoa(port)))

		if err != nil {
			return res.Success(values.NewString(constants.STR_ERR)) // Address resolution failed
		}

		conn, err := net.DialUDP("udp", nil, raddr)

		if err != nil {
			return res.Success(values.NewString(constants.STR_ERR))
		}

		socket.UdpConn = conn
		socket.Address = net.JoinHostPort(address, strconv.Itoa(port))

	case "listen":
		// TCP server (bind & listen)
		listener, err := net.Listen("tcp", net.JoinHostPort(address, strconv.Itoa(port)))

		if err != nil {
			return res.Success(values.NewString(constants.STR_ERR))
		}

		socket.Listener = listener
	}

	handle := registry.Alloc.Alloc()
	registry.Sockets.Store(handle, socket)

	return res.Success(values.NewNumber(handle))
}
