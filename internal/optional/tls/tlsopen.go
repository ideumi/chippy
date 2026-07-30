/*
 *
 * RR2 - internal/optional/tls/tlsopen.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"crypto/tls"
	"net"
	"strconv"
)

func tlsopenFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "open"), 3, "host, port, verify"))
	}

	hostStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "open"), shared.PositionFirst, shared.TypeString, "host"))
	}

	portNum := args[1]

	if !portNum.IsNumber() {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "open"), shared.PositionSecond, shared.TypeNumber, "port"))
	}

	port64, err := portNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	port := int(port64)

	if port < 1 || port > 65535 {
		return res.FailAt(2, shared.Errors.InvalidValue("Port must be between 1 and 65535"))
	}

	verify := true

	if str, ok := values.AsString(args[2]); ok {
		if str.Value == "false" || str.Value == "" {
			verify = false
		}
	} else if args[2].IsNumber() {
		if !args[2].IsTrue() {
			verify = false
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !verify,
		ServerName:         hostStr.Value,
	}

	address := net.JoinHostPort(hostStr.Value, strconv.Itoa(port))
	conn, err := tls.Dial("tcp", address, tlsConfig)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	handle := getNextTLSHandle(ctx)

	storeTLSHandle(ctx, handle, &TLSHandle{
		Conn:   conn,
		Mode:   "client",
		Closed: false,
	})

	return res.Success(values.NewNumber(handle))
}
