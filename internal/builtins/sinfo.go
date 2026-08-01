/*
 *
 * Chippy - internal/builtins/sinfo.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"net"
	"strconv"
)

func sinfoFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sinfo", 1, "handle"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("sinfo", shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	socket, exists := registry.Sockets.Get(handle)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid socket handle"))
	}

	var localIp, remoteIp string
	var localPort, remotePort int

	switch socket.Mode {

	case "tcp":
		if socket.Conn != nil {
			localIp, localPort = addrToIPPort(socket.Conn.LocalAddr())
			remoteIp, remotePort = addrToIPPort(socket.Conn.RemoteAddr())
		}

	case "udp":
		if socket.UdpConn != nil {
			localIp, localPort = addrToIPPort(socket.UdpConn.LocalAddr())
			remoteIp, remotePort = addrToIPPort(socket.UdpConn.RemoteAddr())
		}

	case "listen":
		if socket.Listener != nil {
			localIp, localPort = addrToIPPort(socket.Listener.Addr())
		}
	}

	keys := []string{
		"mode",
		"localIp",
		"localPort",
		"remoteIp",
		"remotePort",
	}

	entries := map[string]values.Value{
		"mode":       values.NewString(socket.Mode),
		"localIp":    values.NewString(localIp),
		"localPort":  values.NewNumber(localPort),
		"remoteIp":   values.NewString(remoteIp),
		"remotePort": values.NewNumber(remotePort),
	}

	result := values.NewMapFromEntries(keys, entries)

	return res.Success(result)
}

func addrToIPPort(addr net.Addr) (string, int) {
	if addr == nil {
		return "", 0
	}

	switch typed := addr.(type) {

	case *net.TCPAddr:
		return typed.IP.String(), typed.Port

	case *net.UDPAddr:
		return typed.IP.String(), typed.Port
	}

	host, portStr, err := net.SplitHostPort(addr.String())

	if err != nil {
		return addr.String(), 0
	}

	port, err := strconv.Atoi(portStr)

	if err != nil {
		return host, 0
	}

	return host, port
}
