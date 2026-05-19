/*
 *
 * RR2 - internal/optional/tls/tlsinfo.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"net"
	"strconv"
)

func tlsinfoFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "info"), 1, "handle")))
	}

	handleNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(
				optional.Prefixed(OptionalName, "info"), shared.TypeNumber, "handle")))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)
	tlsHandle, ok := getTLSHandle(ctx, handle)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Invalid TLS handle")))
	}

	if tlsHandle.Closed || tlsHandle.Conn == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("TLS connection is closed")))
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
		"mode":       values.NewString("tls-" + tlsHandle.Mode).SetContext(ctx),
		"localIp":    values.NewString(localIp).SetContext(ctx),
		"localPort":  values.NewNumber(localPort).SetContext(ctx),
		"remoteIp":   values.NewString(remoteIp).SetContext(ctx),
		"remotePort": values.NewNumber(remotePort).SetContext(ctx),
		"serverName": values.NewString(state.ServerName).SetContext(ctx),
	}

	result := values.NewMapFromEntries(keys, entries)

	return res.Success(result.SetContext(ctx))
}

func tlsAddrToIPPort(addr net.Addr) (string, int) {
	if addr == nil {
		return "", 0
	}

	switch a := addr.(type) {

	case *net.TCPAddr:
		return a.IP.String(), a.Port

	case *net.UDPAddr:
		return a.IP.String(), a.Port
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
