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
	"chip-go/internal/values"
)

func sacceptFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("saccept", 1, "serverHandle"),
			ctx,
		))
	}

	// Get server handle argument
	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("saccept", shared.PositionFirst, shared.TypeNumber, "serverHandle"),
			ctx,
		))
	}

	serverHandle := int(handleNum.Value)

	// Get server socket handle
	serverSocket, exists := shared.GetSocketHandle(serverHandle)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid server socket handle"),
			ctx,
		))
	}

	// Verify this is a listening socket
	if serverSocket.Mode != "listen" {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Socket is not in listening mode. Use sopen() with 'listen' mode first"),
			ctx,
		))
	}

	if serverSocket.Listener == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Listening socket is closed"),
			ctx,
		))
	}

	// Accept incoming connection
	conn, err := serverSocket.Listener.Accept()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Create new socket handle for the accepted connection
	clientHandle := shared.GetNextSocketHandle()

	clientSocket := &shared.SocketHandle{
		Conn: conn,
		Mode: "tcp", // Accepted connections are always TCP
	}

	shared.StoreSocketHandle(clientHandle, clientSocket)

	return res.Success(values.NewNumber(float64(clientHandle)).SetContext(ctx))
}
