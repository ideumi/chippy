/*
 *
 * Modena - internal/context/schema.go
 *
 */

package context

type Schema struct {
	byName map[string]int
	names  []string
}

func NewSchema() *Schema {
	return &Schema{byName: make(map[string]int)}
}

func (s *Schema) Intern(name string) int {
	if slot, ok := s.byName[name]; ok {
		return slot
	}

	slot := len(s.names)
	s.names = append(s.names, name)
	s.byName[name] = slot

	return slot
}

func (s *Schema) SlotOf(name string) (int, bool) {
	slot, ok := s.byName[name]

	return slot, ok
}

func (s *Schema) NameOf(slot int) string {
	if slot < 0 || slot >= len(s.names) {
		return ""
	}

	return s.names[slot]
}
