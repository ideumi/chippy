/*
 *
 * Chippy - internal/optional/json/init.go
 *
 */

package json

import (
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

const OptionalName = "json"

func init() {
	optional.RegisterFactory(OptionalName, GetSection)
}

func GetSection() (map[string]values.Value, map[string]values.Value) {
	return getFunctions(), getConstants()
}

func getFunctions() map[string]values.Value {
	return map[string]values.Value{

		optional.Prefixed(OptionalName, "parse"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "parse"), parseFunction),

		optional.Prefixed(OptionalName, "stringify"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "stringify"), stringifyFunction),
	}
}

func getConstants() map[string]values.Value {
	return map[string]values.Value{
		optional.Prefixed(OptionalName, "TRUE"): values.NewString(jsonTrue),

		optional.Prefixed(OptionalName, "FALSE"): values.NewString(jsonFalse),

		optional.Prefixed(OptionalName, "NULL"): values.NewString(jsonNull),
	}
}
