/*
 *
 * RR2 - internal/optional/http/parseresponse.go
 *
 */

package http

import (
	"bytes"
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"strconv"
	"strings"
)

func parseresponseFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "parseresponse"), 1, "responseBytes"),
			ctx,
		))
	}

	responseBytes, ok := args[0].(*values.Bytes)
	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(
				optional.Prefixed(OptionalName, "parseresponse"), shared.TypeBytes, "responseBytes"),
			ctx,
		))
	}

	data := responseBytes.Data

	// Find end of headers
	headerEnd := bytes.Index(data, []byte("\r\n\r\n"))

	if headerEnd == -1 {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	headerSection := string(data[:headerEnd])
	bodyStart := headerEnd + 4

	// Parse status line
	lines := strings.Split(headerSection, "\r\n")

	if len(lines) == 0 {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	statusLine := lines[0]

	parts := strings.SplitN(statusLine, " ", 3)

	if len(parts) < 2 {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	statusCode, err := strconv.Atoi(parts[1])

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	/* status like required in RFC 9110/9112:
	 * https://datatracker.ietf.org/doc/html/rfc9110 (15.)
	 * 	See libhttp.chh "HttpStatusText(statusCode)"
	 * https://datatracker.ietf.org/doc/html/rfc9112 (4.)
	 *	status-code    = 3DIGIT
	 *	status-line = HTTP-version SP status-code SP [ reason-phrase ]
	 */
	statusText := ""

	if len(parts) >= 3 {
		statusText = parts[2]
	}

	// Parse headers
	headersList := []values.Value{}
	contentLength := -1
	isChunked := false

	for i := 1; i < len(lines); i++ {
		line := lines[i]

		if line == "" {
			continue
		}

		colonIdx := strings.Index(line, ":")

		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:colonIdx])
		value := strings.TrimSpace(line[colonIdx+1:])

		headersList = append(headersList, values.NewList([]values.Value{
			values.NewString(key).SetContext(ctx),
			values.NewString(value).SetContext(ctx),
		}).SetContext(ctx))

		keyLower := strings.ToLower(key)

		if keyLower == "content-length" {
			contentLength, _ = strconv.Atoi(value)
		}

		if keyLower == "transfer-encoding" && strings.Contains(strings.ToLower(value), "chunked") {
			isChunked = true
		}
	}

	// Extract body
	var body []byte

	if bodyStart < len(data) {
		if contentLength >= 0 && !isChunked {
			/* Use Content-Length
			   RFC 9112 6.3:
			   "If a message is received with both a Transfer-Encoding
			   and a Content-Length header field, the Transfer-Encoding
			   overrides the Content-Length."
			*/
			end := bodyStart + contentLength

			if end > len(data) {
				end = len(data)
			}

			body = data[bodyStart:end]
		} else {
			// Chunked or unknown length: take remaining data
			body = data[bodyStart:]
		}
	}

	// Response map
	mapElements := []values.Value{
		values.NewString("map").SetContext(ctx),
	}

	// Status code
	mapElements = append(mapElements, values.NewList([]values.Value{
		values.NewString("statusCode").SetContext(ctx),
		values.NewNumber(float64(statusCode)).SetContext(ctx),
	}).SetContext(ctx))

	// Status text
	mapElements = append(mapElements, values.NewList([]values.Value{
		values.NewString("statusText").SetContext(ctx),
		values.NewString(statusText).SetContext(ctx),
	}).SetContext(ctx))

	// Headers
	mapElements = append(mapElements, values.NewList([]values.Value{
		values.NewString("headers").SetContext(ctx),
		values.NewList(headersList).SetContext(ctx),
	}).SetContext(ctx))

	// Body
	mapElements = append(mapElements, values.NewList([]values.Value{
		values.NewString("body").SetContext(ctx),
		values.NewBytes(body).SetContext(ctx),
	}).SetContext(ctx))

	// Chunked
	mapElements = append(mapElements, values.NewList([]values.Value{
		values.NewString("chunked").SetContext(ctx),
		values.NewNumber(boolToFloat(isChunked)).SetContext(ctx),
	}).SetContext(ctx))

	return res.Success(values.NewList(mapElements).SetContext(ctx))
}

func boolToFloat(b bool) float64 {
	if b {
		return constants.NUM_TRU
	}

	return constants.NUM_FAL
}
