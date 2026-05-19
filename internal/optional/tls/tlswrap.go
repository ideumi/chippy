/*
 *
 * RR2 - internal/optional/tls/tlswrap.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"crypto/tls"
)

func tlswrapFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(
				optional.Prefixed(OptionalName, "wrap"), 3, "socketHandle, certPath, keyPath")))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "wrap"), shared.PositionFirst, shared.TypeNumber, "socketHandle")))
	}

	certPathStr, ok := args[1].(*values.String)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "wrap"), shared.PositionSecond, shared.TypeString, "certPath")))
	}

	keyPathStr, ok := args[2].(*values.String)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "wrap"), shared.PositionThird, shared.TypeString, "keyPath")))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	socketHandle := int(handle64)
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	socket, exists := registry.Sockets.Get(socketHandle)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid socket handle")))
	}

	if socket.Mode != "tcp" {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Socket is not in tcp mode. Use saccept() to obtain a TCP client socket first")))
	}

	if socket.Conn == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Socket connection is closed")))
	}

	cert, err := tls.LoadX509KeyPair(certPathStr.Value, keyPathStr.Value)

	if err != nil {
		socket.Conn.Close()
		registry.Sockets.Remove(socketHandle)
		registry.Alloc.Free(socketHandle)

		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	tlsConn := tls.Server(socket.Conn, tlsConfig)

	if err := tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		registry.Sockets.Remove(socketHandle)
		registry.Alloc.Free(socketHandle)

		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	registry.Sockets.Remove(socketHandle)
	registry.Alloc.Free(socketHandle)

	clientHandle := getNextTLSHandle(ctx)

	storeTLSHandle(ctx, clientHandle, &TLSHandle{
		Conn:   tlsConn,
		Mode:   "server",
		Closed: false,
	})

	return res.Success(values.NewNumber(clientHandle).SetContext(ctx))
}
