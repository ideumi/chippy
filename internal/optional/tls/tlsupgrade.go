/*
 *
 * Chippy - internal/optional/tls/tlsupgrade.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"crypto/tls"
)

func tlsupgradeFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(
				optional.Prefixed(OptionalName, "upgrade"), 3, "socketHandle, certPath, keyPath"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "upgrade"), shared.PositionFirst, shared.TypeNumber, "socketHandle"))
	}

	certPathStr, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "upgrade"), shared.PositionSecond, shared.TypeString, "certPath"))
	}

	keyPathStr, ok := values.AsString(args[2])

	if !ok {
		return res.FailAt(3,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "upgrade"), shared.PositionThird, shared.TypeString, "keyPath"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	socketHandle := int(handle64)
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	socket, exists := registry.Sockets.Get(socketHandle)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid socket handle"))
	}

	if socket.Mode != "tcp" {
		return res.FailAt(1,
			shared.Errors.InvalidValue("Socket is not in tcp mode. Use saccept() to obtain a TCP client socket first"))
	}

	if socket.Conn == nil {
		return res.FailAt(1, shared.Errors.InvalidValue("Socket connection is closed"))
	}

	cert, err := tls.LoadX509KeyPair(certPathStr.Value, keyPathStr.Value)

	if err != nil {
		socket.Conn.Close()
		registry.Sockets.Remove(socketHandle)
		registry.Alloc.Free(socketHandle)

		return res.Success(values.NewString(constants.STR_ERR))
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConn := tls.Server(socket.Conn, tlsConfig)

	if err := tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		registry.Sockets.Remove(socketHandle)
		registry.Alloc.Free(socketHandle)

		return res.Success(values.NewString(constants.STR_ERR))
	}

	registry.Sockets.Remove(socketHandle)
	registry.Alloc.Free(socketHandle)

	clientHandle := getNextTLSHandle(ctx)

	storeTLSHandle(ctx, clientHandle, &TLSHandle{
		Conn:   tlsConn,
		Mode:   "server",
		Closed: false,
	})

	return res.Success(values.NewNumber(clientHandle))
}
