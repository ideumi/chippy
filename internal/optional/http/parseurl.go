/*
 *
 * Chippy - internal/optional/http/parseurl.go
 *
 */

package http

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"net/url"
	"strconv"
)

func parseurlFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "parseurl"), 1, "url"))
	}

	urlStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint(
				optional.Prefixed(OptionalName, "parseurl"), shared.TypeString, "url"))
	}

	parsedURL, err := url.Parse(urlStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	// Require scheme and host for absolute URLs
	// Empty URLs or URLs without scheme/host should return error
	if urlStr.Value == "" || (parsedURL.Scheme == "" && parsedURL.Host == "") {
		return res.Success(values.NewString(constants.STR_ERR))
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
			"scheme": values.NewString(parsedURL.Scheme),
			"host":   values.NewString(parsedURL.Hostname()),
			"port":   values.NewNumber(portNum),
			"path":   values.NewString(path),
			"query":  values.NewString(parsedURL.RawQuery),
		},
	)

	return res.Success(nativeMap)
}
