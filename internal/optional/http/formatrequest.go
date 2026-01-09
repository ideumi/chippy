/*
 *
 * RR2 - internal/optional/http/formatrequest.go
 *
 */

package http

import (
	"bytes"
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"strconv"
	"strings"
)

func formatrequestFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 4 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "formatrequest"), 4, "method, path, headers, body"),
			ctx,
		))
	}

	methodStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "formatrequest"), shared.PositionFirst, shared.TypeString, "method"),
			ctx,
		))
	}

	pathStr, ok := args[1].(*values.String)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "formatrequest"), shared.PositionSecond, shared.TypeString, "path"),
			ctx,
		))
	}

	headersList, ok := args[2].(*values.List)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "formatrequest"), shared.PositionThird, shared.TypeList, "headers"),
			ctx,
		))
	}

	bodyBytes, ok := args[3].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[3].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "formatrequest"), shared.PositionFourth, shared.TypeBytes, "body"),
			ctx,
		))
	}

	// Build request
	var buf bytes.Buffer

	// Request line
	method := strings.ToUpper(methodStr.Value)
	path := pathStr.Value

	if path == "" {
		path = "/"
	}

	buf.WriteString(method)
	buf.WriteString(" ")
	buf.WriteString(path)
	buf.WriteString(" HTTP/1.1\r\n")

	// Headers
	hasContentLength := false

	for _, elem := range headersList.Elements {
		pair, ok := elem.(*values.List)

		if !ok || len(pair.Elements) != 2 {
			posStart, posEnd := args[2].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Headers must be a list of [key, value] pairs"),
				ctx,
			))
		}

		key, ok1 := pair.Elements[0].(*values.String)
		val, ok2 := pair.Elements[1].(*values.String)

		if !ok1 || !ok2 {
			posStart, posEnd := args[2].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Header keys and values must be strings"),
				ctx,
			))
		}

		keyLower := strings.ToLower(key.Value)

		if keyLower == "content-length" {
			hasContentLength = true
		}

		buf.WriteString(key.Value)
		buf.WriteString(": ")
		buf.WriteString(val.Value)
		buf.WriteString("\r\n")
	}

	// Add Content-Length if body present and not already set
	if len(bodyBytes.Data) > 0 && !hasContentLength {
		buf.WriteString("Content-Length: ")
		buf.WriteString(strconv.Itoa(len(bodyBytes.Data)))
		buf.WriteString("\r\n")
	}

	// End headers
	buf.WriteString("\r\n")

	// Body
	if len(bodyBytes.Data) > 0 {
		buf.Write(bodyBytes.Data)
	}

	return res.Success(values.NewBytes(buf.Bytes()).SetContext(ctx))
}
