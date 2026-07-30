/*
 *
 * RR2 - internal/optional/hash/init.go
 *
 */

package hash

import (
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

const OptionalName = "hash"

func init() {
	optional.RegisterFactory(OptionalName, GetSection)
}

func GetSection() (map[string]values.Value, map[string]values.Value) {
	return getFunctions(), getConstants()
}

func getFunctions() map[string]values.Value {
	return map[string]values.Value{
		optional.Prefixed(OptionalName, "md5"):     values.NewBuiltInFunction(optional.Prefixed(OptionalName, "md5"), md5Function),
		optional.Prefixed(OptionalName, "sha1"):    values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha1"), sha1Function),
		optional.Prefixed(OptionalName, "sha224"):  values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha224"), sha224Function),
		optional.Prefixed(OptionalName, "sha256"):  values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha256"), sha256Function),
		optional.Prefixed(OptionalName, "sha384"):  values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha384"), sha384Function),
		optional.Prefixed(OptionalName, "sha512"):  values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha512"), sha512Function),
		optional.Prefixed(OptionalName, "sha3224"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha3224"), sha3224Function),
		optional.Prefixed(OptionalName, "sha3256"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha3256"), sha3256Function),
		optional.Prefixed(OptionalName, "sha3384"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha3384"), sha3384Function),
		optional.Prefixed(OptionalName, "sha3512"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "sha3512"), sha3512Function),
	}
}

func getConstants() map[string]values.Value {
	return map[string]values.Value{}
}
