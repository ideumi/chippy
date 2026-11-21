/*
 *
 * Chippy - internal/optional/base64/init.go
 *
 */

package base64

import (
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

const OptionalName = "base64"

func init() {
	optional.RegisterFactory(OptionalName, GetSection)
}

func GetSection() (map[string]values.Value, map[string]values.Value) {
	return getFunctions(), getConstants()
}

func getFunctions() map[string]values.Value {
	return map[string]values.Value{
		optional.Prefixed(OptionalName, "encode"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "encode"), base64encodeFunction),
		optional.Prefixed(OptionalName, "decode"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "decode"), base64decodeFunction),
	}
}

func getConstants() map[string]values.Value {
	return map[string]values.Value{}
}
