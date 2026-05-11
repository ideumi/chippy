/*
 *
 * RR2 - internal/optional/tls/tlsopen.go
 *
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"crypto/tls"
	"fmt"
)

func tlsopenFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "open"), 3, "host, port, verify")))
	}

	hostStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "open"), shared.PositionFirst, shared.TypeString, "host")))
	}

	portNum, ok := args[1].(*values.Number)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "open"), shared.PositionSecond, shared.TypeNumber, "port")))
	}

	port := int(portNum.Value)

	if port < 1 || port > 65535 {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Port must be between 1 and 65535")))
	}

	verify := true

	switch v := args[2].(type) {
	case *values.String:
		if v.Value == "false" || v.Value == "" {
			verify = false
		}
	case *values.Number:
		if v.Value == constants.NUM_FAL {
			verify = false
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !verify,
		ServerName:         hostStr.Value,
	}

	address := fmt.Sprintf("%s:%d", hostStr.Value, port)
	conn, err := tls.Dial("tcp", address, tlsConfig)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	handle := getNextTLSHandle(ctx)

	storeTLSHandle(ctx, handle, &TLSHandle{
		Conn:   conn,
		Mode:   "client",
		Closed: false,
	})

	return res.Success(values.NewNumber(float64(handle)).SetContext(ctx))
}
