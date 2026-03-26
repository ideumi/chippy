/*
 *
 * RR2 - internal/optional/http/parseurl.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"net/url"
	"strconv"
)

func parseurlFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "parseurl"), 1, "url"),
			ctx,
		))
	}

	urlStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint(
				optional.Prefixed(OptionalName, "parseurl"), shared.TypeString, "url"),
			ctx,
		))
	}

	parsedURL, err := url.Parse(urlStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Require scheme and host for absolute URLs
	// Empty URLs or URLs without scheme/host should return error
	if urlStr.Value == "" || (parsedURL.Scheme == "" && parsedURL.Host == "") {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Determine port
	port := parsedURL.Port()

	if port == "" {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else if parsedURL.Scheme == "http" {
			port = "80"
		} else {
			port = "0"
		}
	}

	portNum, _ := strconv.Atoi(port)

	// Build path with query
	path := parsedURL.Path

	if path == "" {
		path = "/"
	}

	nativeMap := values.NewMapFromEntries(
		[]string{"scheme", "host", "port", "path", "query"},
		map[string]values.Value{
			"scheme": values.NewString(parsedURL.Scheme).SetContext(ctx),
			"host":   values.NewString(parsedURL.Hostname()).SetContext(ctx),
			"port":   values.NewNumber(float64(portNum)).SetContext(ctx),
			"path":   values.NewString(path).SetContext(ctx),
			"query":  values.NewString(parsedURL.RawQuery).SetContext(ctx),
		},
	)

	return res.Success(nativeMap.SetContext(ctx))
}
