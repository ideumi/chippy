/*
 *
 * Chippy - internal/optional/http/init.go
 *
 */

package http

import (
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

const OptionalName = "http"

func init() {
	optional.RegisterFactory(OptionalName, GetSection)
}

func GetSection() (map[string]values.Value, map[string]values.Value) {
	return getFunctions(), getConstants()
}

func getFunctions() map[string]values.Value {
	return map[string]values.Value{
		optional.Prefixed(OptionalName, "parseurl"):      values.NewBuiltInFunction(optional.Prefixed(OptionalName, "parseurl"), parseurlFunction),
		optional.Prefixed(OptionalName, "formatrequest"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "formatrequest"), formatrequestFunction),
		optional.Prefixed(OptionalName, "parseresponse"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "parseresponse"), parseresponseFunction),
		optional.Prefixed(OptionalName, "dechunk"):       values.NewBuiltInFunction(optional.Prefixed(OptionalName, "dechunk"), dechunkFunction),

		optional.Prefixed(OptionalName, "urlencode"):    values.NewBuiltInFunction(optional.Prefixed(OptionalName, "urlencode"), urlencodeFunction),
		optional.Prefixed(OptionalName, "urldecode"):    values.NewBuiltInFunction(optional.Prefixed(OptionalName, "urldecode"), urldecodeFunction),
		optional.Prefixed(OptionalName, "base64encode"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "base64encode"), base64encodeFunction),
		optional.Prefixed(OptionalName, "base64decode"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "base64decode"), base64decodeFunction),
	}
}

func getConstants() map[string]values.Value {
	return map[string]values.Value{}
}
