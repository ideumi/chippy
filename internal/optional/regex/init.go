/*
 *
 * Chippy - internal/optional/regex/init.go
 *
 */

package regex

import (
	"chip-go/internal/optional"
	"chip-go/internal/values"
)

const OptionalName = "regex"

func init() {
	optional.RegisterFactory(OptionalName, GetSection)
}

func GetSection() (map[string]values.Value, map[string]values.Value) {
	return getFunctions(), getConstants()
}

func getFunctions() map[string]values.Value {
	return map[string]values.Value{
		optional.Prefixed(OptionalName, "match"):          values.NewBuiltInFunction(optional.Prefixed(OptionalName, "match"), matchFunction),
		optional.Prefixed(OptionalName, "find"):           values.NewBuiltInFunction(optional.Prefixed(OptionalName, "find"), findFunction),
		optional.Prefixed(OptionalName, "findindex"):      values.NewBuiltInFunction(optional.Prefixed(OptionalName, "findindex"), findIndexFunction),
		optional.Prefixed(OptionalName, "findall"):        values.NewBuiltInFunction(optional.Prefixed(OptionalName, "findall"), findAllFunction),
		optional.Prefixed(OptionalName, "findallindexes"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "findallindexes"), findAllIndexesFunction),
		optional.Prefixed(OptionalName, "findallgroups"):  values.NewBuiltInFunction(optional.Prefixed(OptionalName, "findallgroups"), findAllGroupsFunction),
		optional.Prefixed(OptionalName, "groups"):         values.NewBuiltInFunction(optional.Prefixed(OptionalName, "groups"), groupsFunction),
		optional.Prefixed(OptionalName, "groupnames"):     values.NewBuiltInFunction(optional.Prefixed(OptionalName, "groupnames"), groupNamesFunction),
		optional.Prefixed(OptionalName, "groupindex"):     values.NewBuiltInFunction(optional.Prefixed(OptionalName, "groupindex"), groupIndexFunction),
		optional.Prefixed(OptionalName, "replace"):        values.NewBuiltInFunction(optional.Prefixed(OptionalName, "replace"), replaceFunction),
		optional.Prefixed(OptionalName, "replacefirst"):   values.NewBuiltInFunction(optional.Prefixed(OptionalName, "replacefirst"), replaceFirstFunction),
		optional.Prefixed(OptionalName, "replaceliteral"): values.NewBuiltInFunction(optional.Prefixed(OptionalName, "replaceliteral"), replaceLiteralFunction),
		optional.Prefixed(OptionalName, "escape"):         values.NewBuiltInFunction(optional.Prefixed(OptionalName, "escape"), escapeFunction),
		optional.Prefixed(OptionalName, "split"):          values.NewBuiltInFunction(optional.Prefixed(OptionalName, "split"), splitFunction),
		optional.Prefixed(OptionalName, "valid"):          values.NewBuiltInFunction(optional.Prefixed(OptionalName, "valid"), validFunction),
	}
}

func getConstants() map[string]values.Value {
	return map[string]values.Value{}
}
