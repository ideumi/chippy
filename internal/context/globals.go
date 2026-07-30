/*
 *
 * Chippy - internal/context/globals.go
 *
 */

package context

type Globals[T Storable] struct {
	schema *Schema
	values []T
}

func NewGlobals[T Storable](schema *Schema) *Globals[T] {
	return &Globals[T]{schema: schema}
}

func (g *Globals[T]) Schema() *Schema {
	return g.schema
}

func (g *Globals[T]) grow(size int) {
	if size <= len(g.values) {
		return
	}

	g.values = append(g.values, make([]T, size-len(g.values))...)
}

func (g *Globals[T]) SlotValue(slot int) T {
	if slot < 0 || slot >= len(g.values) {
		var missing T

		return missing
	}

	return g.values[slot]
}

func (g *Globals[T]) SetSlot(slot int, value T) {
	g.grow(slot + 1)
	g.values[slot] = value
}

func (g *Globals[T]) GetByName(name string) T {
	if slot, ok := g.schema.SlotOf(name); ok {
		return g.SlotValue(slot)
	}

	var missing T

	return missing
}

func (g *Globals[T]) SetByName(name string, value T) {
	g.SetSlot(g.schema.Intern(name), value)
}

func (g *Globals[T]) ForEach(fn func(name string, value T)) {
	for slot := 0; slot < len(g.values); slot++ {
		if g.values[slot].IsSet() {
			fn(g.schema.NameOf(slot), g.values[slot])
		}
	}
}
