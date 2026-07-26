/*
 *
 * Modena - internal/globals/schema.go
 *
 */

package globals

import "sync"

type Schema struct {
	mu     sync.RWMutex
	byName map[string]int
	names  []string
}

func NewSchema() *Schema {
	return &Schema{byName: make(map[string]int)}
}

func (s *Schema) Intern(name string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if slot, ok := s.byName[name]; ok {
		return slot
	}

	slot := len(s.names)
	s.names = append(s.names, name)
	s.byName[name] = slot

	return slot
}

func (s *Schema) SlotOf(name string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	slot, ok := s.byName[name]

	return slot, ok
}

func (s *Schema) NameOf(slot int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if slot < 0 || slot >= len(s.names) {
		return ""
	}

	return s.names[slot]
}
