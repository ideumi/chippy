/*
 *
 * Chippy - internal/builtins/saccept.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/handles"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
)

func sacceptFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("saccept", 1, "serverHandle"))
	}

	// Get server handle argument
	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("saccept", shared.TypeNumber, "serverHandle"))
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
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid server socket handle"))
	}

	// Verify this is a listening socket
	if serverSocket.Mode != "tcplisten" && serverSocket.Mode != "unixlisten" {
		return res.FailAt(1,
			shared.Errors.InvalidValue("Socket is not in listening mode. Use sopen() with 'tcplisten' or 'unixlisten' mode first"))
	}

	if serverSocket.Listener == nil {
		return res.FailAt(1, shared.Errors.InvalidValue("Listening socket is closed"))
	}

	// Accept incoming connection
	conn, err := serverSocket.Listener.Accept()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	// Accepted connections take the family of the listener
	var mode string

	switch serverSocket.Mode {

	case "tcplisten":
		mode = "tcp"

	case "unixlisten":
		mode = "unix"

	default:
		conn.Close()
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid listener mode"))
	}

	// Create new socket handle for the accepted connection
	clientHandle := registry.Alloc.Alloc()

	clientSocket := &handles.SocketHandle{
		Conn: conn,
		Mode: mode,
	}

	registry.Sockets.Store(clientHandle, clientSocket)

	return res.Success(values.NewNumber(clientHandle))
}
