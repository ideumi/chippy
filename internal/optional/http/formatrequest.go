/*
 *
 * Chippy - internal/optional/http/formatrequest.go
 *
 */

package http

import (
	"bytes"
	"chip-go/internal/builtins/shared"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"strconv"
	"strings"
)

func formatrequestFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 4 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "formatrequest"), 4, "method, path, headers, body"))
	}

	methodStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "formatrequest"), shared.PositionFirst, shared.TypeString, "method"))
	}

	pathStr, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "formatrequest"), shared.PositionSecond, shared.TypeString, "path"))
	}

	headersList, ok := values.AsList(args[2])

	if !ok {
		return res.FailAt(3,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "formatrequest"), shared.PositionThird, shared.TypeList, "headers"))
	}

	bodyBytes, ok := values.AsBytes(args[3])

	if !ok {
		return res.FailAt(4,
			shared.Errors.InvalidArgTypePositionalWithHint(
				optional.Prefixed(OptionalName, "formatrequest"), shared.PositionFourth, shared.TypeBytes, "body"))
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
	hasConnection := false

	for _, elem := range headersList.Elements {
		pair, ok := values.AsList(elem)

		if !ok || len(pair.Elements) != 2 {
			return res.FailAt(3,
				shared.Errors.InvalidValue("Headers must be a list of [key, value] pairs"))
		}

		key, ok1 := values.AsString(pair.Elements[0])
		val, ok2 := values.AsString(pair.Elements[1])

		if !ok1 || !ok2 {
			return res.FailAt(3, shared.Errors.InvalidValue("Header keys and values must be strings"))
		}

		keyLower := strings.ToLower(key.Value)

		if keyLower == "content-length" {
			hasContentLength = true
		}

		if keyLower == "connection" {
			hasConnection = true
		}

		buf.WriteString(key.Value)
		buf.WriteString(": ")
		buf.WriteString(val.Value)
		buf.WriteString("\r\n")
	}

	// Add Content-Length if not already set
	if !hasContentLength {
		buf.WriteString("Content-Length: ")
		buf.WriteString(strconv.Itoa(len(bodyBytes.Data)))
		buf.WriteString("\r\n")
	}

	// Close if not already set
	if !hasConnection {
		buf.WriteString("Connection: close\r\n")
	}

	// End headers
	buf.WriteString("\r\n")

	// Body
	if len(bodyBytes.Data) > 0 {
		buf.Write(bodyBytes.Data)
	}

	return res.Success(values.NewBytes(buf.Bytes()))
}
