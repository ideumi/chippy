/*
 *
 * Chippy - internal/optional/registry.go
 *
 */

package optional

import (
	"chip-go/internal/values"
	"fmt"
)

type Optional struct {
	Name      string
	Functions map[string]values.Value
	Constants map[string]values.Value
}

type OptionalFactory func() (map[string]values.Value, map[string]values.Value)

var optionalFactories = make(map[string]OptionalFactory)

func Prefixed(optional, name string) string {
	return fmt.Sprintf("_%s_%s_", optional, name)
}

func RegisterFactory(name string, factory OptionalFactory) {
	optionalFactories[name] = factory
}

func GetOptional(name string) (*Optional, bool) {
	factory, exists := optionalFactories[name]

	if !exists {
		return nil, false
	}

	// Lazy init
	funcs, consts := factory()

	return &Optional{
		Name:      name,
		Functions: funcs,
		Constants: consts,
	}, true
}
