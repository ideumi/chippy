/*
 *
 * Chippy - internal/optional/tls/init.go
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

func GetSection() (map[string]values.Value, map[string]values.Value) {
	return getFunctions(), getConstants()
}

func getFunctions() map[string]values.Value {
	return map[string]values.Value{
		optional.Prefixed(OptionalName, "open"):    values.NewBuiltInFunction(optional.Prefixed(OptionalName, "open"), tlsopenFunction),
		optional.Prefixed(OptionalName, "read"):    values.NewBuiltInFunction(optional.Prefixed(OptionalName, "read"), tlsreadFunction),
		optional.Prefixed(OptionalName, "write"):   values.NewBuiltInFunction(optional.Prefixed(OptionalName, "write"), tlswriteFunction),
		optional.Prefixed(OptionalName, "close"):   values.NewBuiltInFunction(optional.Prefixed(OptionalName, "close"), tlscloseFunction),
		optional.Prefixed(OptionalName, "upgrade"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "upgrade"), tlsupgradeFunction),
		optional.Prefixed(OptionalName, "info"):    values.NewBuiltInFunction(optional.Prefixed(OptionalName, "info"), tlsinfoFunction),
	}
}

func getConstants() map[string]values.Value {
	return map[string]values.Value{}
}
