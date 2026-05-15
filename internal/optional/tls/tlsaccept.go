/*
 *
 * RR2 - internal/optional/tls/tlsaccept.go
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

func tlsacceptFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "accept"), 3, "serverHandle, certPath, keyPath")))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "accept"), shared.PositionFirst, shared.TypeNumber, "serverHandle")))
	}

	certPathStr, ok := args[1].(*values.String)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "accept"), shared.PositionSecond, shared.TypeString, "certPath")))
	}

	keyPathStr, ok := args[2].(*values.String)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "accept"), shared.PositionThird, shared.TypeString, "keyPath")))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	serverHandle := int(handle64)
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	serverSocket, exists := registry.Sockets.Get(serverHandle)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid server socket handle")))
	}

	if serverSocket.Mode != "listen" {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Socket is not in listening mode")))
	}

	if serverSocket.Listener == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Listening socket is closed")))
	}

	cert, err := tls.LoadX509KeyPair(certPathStr.Value, keyPathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	conn, err := serverSocket.Listener.Accept()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	tlsConn := tls.Server(conn, tlsConfig)

	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	clientHandle := getNextTLSHandle(ctx)

	storeTLSHandle(ctx, clientHandle, &TLSHandle{
		Conn:   tlsConn,
		Mode:   "server",
		Closed: false,
	})

	return res.Success(values.NewNumber(clientHandle).SetContext(ctx))
}
