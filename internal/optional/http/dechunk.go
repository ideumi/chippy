/*
 *
 * RR2 - internal/optional/http/dechunk.go
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

func dechunkFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "dechunk"), 1, "chunkedBytes")))
	}

	chunkedBytes, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(
				optional.Prefixed(OptionalName, "dechunk"), shared.TypeBytes, "chunkedBytes")))
	}

	data := chunkedBytes.Data
	var result bytes.Buffer
	pos := 0

	for pos < len(data) {
		// Find end of size line \r\n
		lineEnd := bytes.Index(data[pos:], []byte("\r\n"))

		if lineEnd == -1 {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		// Parse chunk size
		sizeLine := string(data[pos : pos+lineEnd])

		// Strip chunk extensions
		if idx := strings.Index(sizeLine, ";"); idx != -1 {
			sizeLine = sizeLine[:idx]
		}

		sizeLine = strings.TrimSpace(sizeLine)

		chunkSize, err := strconv.ParseInt(sizeLine, 16, 64)

		if err != nil || chunkSize < 0 {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		pos += lineEnd + 2 // Skip size line and \r\n

		// Check for last chunk
		if chunkSize == 0 {
			break
		}

		// Read chunk data and compare against remaining length to avoid overflow
		if chunkSize > int64(len(data)-pos) {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		result.Write(data[pos : pos+int(chunkSize)])
		pos += int(chunkSize)

		// Skip trailing \r\n after chunk data
		if pos+2 > len(data) || data[pos] != '\r' || data[pos+1] != '\n' {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		pos += 2
	}

	return res.Success(values.NewBytes(result.Bytes()).SetContext(ctx))
}
