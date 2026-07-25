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
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"strconv"
	"strings"
)

func parseresponseFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "parseresponse"), 1, "responseBytes"))
	}

	responseBytes, ok := args[0].(*values.Bytes)
	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(
				optional.Prefixed(OptionalName, "parseresponse"), shared.TypeBytes, "responseBytes"))
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
			// Leave contentLength at -1 on garbage so the body defaults to
			// the remaining data rather than being truncated to empty
			if n, err := strconv.Atoi(value); err == nil {
				contentLength = n
			}
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

			// Clamp to remaining length to avoid overflow on a huge Content-Length
			end := len(data)

			if contentLength < len(data)-bodyStart {
				end = bodyStart + contentLength
			}

			body = data[bodyStart:end]
		} else {
			// Chunked or unknown length: take remaining data
			body = data[bodyStart:]
		}
	}

	nativeMap := values.NewMapFromEntries(
		[]string{"statusCode", "statusText", "headers", "body", "chunked"},
		map[string]values.Value{
			"statusCode": values.NewNumber(statusCode).SetContext(ctx),
			"statusText": values.NewString(statusText).SetContext(ctx),
			"headers":    values.NewList(headersList).SetContext(ctx),
			"body":       values.NewBytes(body).SetContext(ctx),
			"chunked":    values.NewNumber(boolToInt(isChunked)).SetContext(ctx),
		},
	)

	return res.Success(nativeMap.SetContext(ctx))
}

func boolToInt(b bool) int {
	if b {
		return constants.NUM_TRU
	}

	return constants.NUM_FAL
}
