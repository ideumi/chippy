/*
 *
 * RR2 - internal/builtins/saccept.go
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
)

func sacceptFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("saccept", 1, "serverHandle")))
	}

	// Get server handle argument
	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("saccept", shared.TypeNumber, "serverHandle")))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	serverHandle := int(handle64)

	// Get server socket handle
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	serverSocket, exists := registry.Sockets.Get(serverHandle)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid server socket handle")))
	}

	// Verify this is a listening socket
	if serverSocket.Mode != "listen" {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Socket is not in listening mode. Use sopen() with 'listen' mode first")))
	}

	if serverSocket.Listener == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Listening socket is closed")))
	}

	// Accept incoming connection
	conn, err := serverSocket.Listener.Accept()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Create new socket handle for the accepted connection
	clientHandle := registry.Alloc.Alloc()

	clientSocket := &handles.SocketHandle{
		Conn: conn,
		Mode: "tcp", // Accepted connections are always TCP
	}

	registry.Sockets.Store(clientHandle, clientSocket)

	return res.Success(values.NewNumber(clientHandle).SetContext(ctx))
}
