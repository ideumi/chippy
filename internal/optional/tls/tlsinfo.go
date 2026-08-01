/*
 *
 * Chippy - internal/optional/tls/tlsinfo.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"net"
	"strconv"
)

func tlsinfoFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "info"), 1, "handle"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(
				optional.Prefixed(OptionalName, "info"), shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)
	tlsHandle, ok := getTLSHandle(ctx, handle)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid TLS handle"))
	}

	if tlsHandle.Closed || tlsHandle.Conn == nil {
		return res.FailAt(1, shared.Errors.InvalidValue("TLS connection is closed"))
	}

	localIp, localPort := tlsAddrToIPPort(tlsHandle.Conn.LocalAddr())
	remoteIp, remotePort := tlsAddrToIPPort(tlsHandle.Conn.RemoteAddr())

	state := tlsHandle.Conn.ConnectionState()

	keys := []string{
		"mode",
		"localIp",
		"localPort",
		"remoteIp",
		"remotePort",
		"serverName",
	}

	entries := map[string]values.Value{
		"mode":       values.NewString("tls-" + tlsHandle.Mode),
		"localIp":    values.NewString(localIp),
		"localPort":  values.NewNumber(localPort),
		"remoteIp":   values.NewString(remoteIp),
		"remotePort": values.NewNumber(remotePort),
		"serverName": values.NewString(state.ServerName),
	}

	result := values.NewMapFromEntries(keys, entries)

	return res.Success(result)
}

func tlsAddrToIPPort(addr net.Addr) (string, int) {
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
