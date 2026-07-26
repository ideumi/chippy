/*
 *
 * Modena - internal/values/globals.go
 *
 */

package values

import "chip-go/internal/globals"

type GlobalStore struct {
	schema *globals.Schema
	values []Value
}

func NewGlobalStore(schema *globals.Schema) *GlobalStore {
	return &GlobalStore{schema: schema}
}

func (g *GlobalStore) grow(size int) {
	if size <= len(g.values) {
		return
	}

	grown := make([]Value, size)
	copy(grown, g.values)
	g.values = grown
}

func (g *GlobalStore) SlotValue(slot int) Value {
	if slot < 0 || slot >= len(g.values) {
		return nil
	}

	return g.values[slot]
}

func (g *GlobalStore) SetSlot(slot int, value Value) {
	g.grow(slot + 1)
	g.values[slot] = value
}

func (g *GlobalStore) GetByName(name string) Value {
	if slot, ok := g.schema.SlotOf(name); ok {
		return g.SlotValue(slot)
	}

	return nil
}

func (g *GlobalStore) SetByName(name string, value Value) {
	g.SetSlot(g.schema.Intern(name), value)
}

func (g *GlobalStore) ForEach(fn func(name string, value Value)) {
	for slot := 0; slot < len(g.values); slot++ {
		if g.values[slot] != nil {
			fn(g.schema.NameOf(slot), g.values[slot])
		}
	}
}
