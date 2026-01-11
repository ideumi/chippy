/*
 *
 * RR2 - internal/optional/tls/init.go
 *
 */

package tls

import (
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

const OptionalName = "tls"

func init() {
	optional.RegisterFactory(OptionalName, GetSection)
}

func GetSection() (map[string]*values.BuiltInFunction, map[string]values.Value) {
	return getFunctions(), getConstants()
}

func getFunctions() map[string]*values.BuiltInFunction {
	return map[string]*values.BuiltInFunction{
		optional.Prefixed(OptionalName, "open"):   values.NewBuiltInFunction(optional.Prefixed(OptionalName, "open"), tlsopenFunction),
		optional.Prefixed(OptionalName, "read"):   values.NewBuiltInFunction(optional.Prefixed(OptionalName, "read"), tlsreadFunction),
		optional.Prefixed(OptionalName, "write"):  values.NewBuiltInFunction(optional.Prefixed(OptionalName, "write"), tlswriteFunction),
		optional.Prefixed(OptionalName, "close"):  values.NewBuiltInFunction(optional.Prefixed(OptionalName, "close"), tlscloseFunction),
		optional.Prefixed(OptionalName, "accept"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "accept"), tlsacceptFunction),
	}
}

func getConstants() map[string]values.Value {
	return map[string]values.Value{}
}
