/*
 *
 * Chippy - internal/values/map.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"strings"
)

type Map struct {
	OperatorDefaults
	Keys    []string
	Entries map[string]Value
}

func NewMapFromEntries(keys []string, entries map[string]Value) Value {
	return fromHeap(TagMap, &Map{Keys: keys, Entries: entries})
}

func (m *Map) wrap() Value {
	return fromHeap(TagMap, m)
}

func (m *Map) String() string {
	return m.stringWalk(map[Value]bool{}, 0)
}

func (m *Map) stringWalk(seen map[Value]bool, depth int) string {
	self := m.wrap()

	if depth > constants.LIMIT_VALUE_NESTING_DEPTH || seen[self] {
		return "m[...]"
	}

	seen[self] = true
	defer delete(seen, self)

	pairs := make([]string, len(m.Keys))

	for i, key := range m.Keys {
		valStr := walkString(m.Entries[key], seen, depth+1)

		pairs[i] = "\"" + key + "\": " + valStr
	}

	return "m[" + strings.Join(pairs, ", ") + "]"
}

func (m *Map) Copy() Value {
	return m.copyWalk(map[Value]bool{}, 0)
}

func (m *Map) copyWalk(seen map[Value]bool, depth int) Value {
	defer EnterWalk(m.wrap(), seen, depth)()

	newKeys := make([]string, len(m.Keys))
	copy(newKeys, m.Keys)
	newEntries := make(map[string]Value, len(m.Entries))

	for key, val := range m.Entries {
		newEntries[key] = walkCopy(val, seen, depth+1)
	}

	return NewMapFromEntries(newKeys, newEntries)
}

func (m *Map) shallowCopy() *Map {
	newKeys := make([]string, len(m.Keys))
	copy(newKeys, m.Keys)

	newEntries := make(map[string]Value, len(m.Entries))

	for key, val := range m.Entries {
		newEntries[key] = val
	}

	return &Map{Keys: newKeys, Entries: newEntries}
}

func (m *Map) IsTrue() bool {
	return len(m.Keys) > 0
}

func (m *Map) Set(key string, val Value) {
	if _, exists := m.Entries[key]; !exists {
		m.Keys = append(m.Keys, key)
	}

	m.Entries[key] = val
}

func (m *Map) mapRemove(key string) (*Map, bool) {
	if _, exists := m.Entries[key]; !exists {
		return nil, false
	}

	newMap := m.shallowCopy()
	delete(newMap.Entries, key)
	newKeys := make([]string, 0, len(newMap.Keys)-1)

	for _, existingKey := range newMap.Keys {
		if existingKey != key {
			newKeys = append(newKeys, existingKey)
		}
	}

	newMap.Keys = newKeys

	return newMap, true
}

func (m *Map) AddedTo(other Value) (Value, error) {
	otherMap, ok := AsMap(other)

	if !ok {
		return Value{}, IllegalOperation()
	}

	result := m.shallowCopy()

	for _, key := range otherMap.Keys {
		result.Set(key, otherMap.Entries[key])
	}

	return result.wrap(), nil
}

func (m *Map) SubbedBy(other Value) (Value, error) {
	otherStr, ok := AsString(other)

	if !ok {
		return Value{}, IllegalOperation()
	}

	newMap, found := m.mapRemove(otherStr.Value)

	if !found {
		return NewString(constants.STR_ERR), nil
	}

	return newMap.wrap(), nil
}

func (m *Map) GetComparisonEe(other Value) (Value, error) {
	return m.eqWalk(other, map[Value]bool{}, 0)
}

func (m *Map) eqWalk(other Value, seen map[Value]bool, depth int) (Value, error) {
	leave, err := enterWalk(m.wrap(), seen, depth)

	if err != nil {
		return Value{}, err
	}

	defer leave()

	otherMap, ok := AsMap(other)

	if !ok {
		return Value{}, IllegalOperation()
	}

	if len(m.Keys) != len(otherMap.Keys) {
		return Bool(false), nil
	}

	for key, val := range m.Entries {
		otherVal, exists := otherMap.Entries[key]

		if !exists {
			return Bool(false), nil
		}

		if val.IsUnset() && otherVal.IsUnset() {
			continue
		}

		if val.IsUnset() || otherVal.IsUnset() {
			return Bool(false), nil
		}

		comparison, err := walkEqual(val, otherVal, seen, depth+1)

		if err != nil {
			return Value{}, err
		}

		if !comparison.IsTrue() {
			return Bool(false), nil
		}
	}

	return Bool(true), nil
}

func (m *Map) GetComparisonNe(other Value) (Value, error) {
	comparison, err := m.GetComparisonEe(other)

	if err != nil {
		return Value{}, err
	}

	return Bool(!comparison.IsTrue()), nil
}

func (m *Map) Notted() (Value, error) {
	return Bool(!m.IsTrue()), nil
}

func (m *Map) XoredBy(other Value) (Value, error) {
	return Bool(m.IsTrue() != other.IsTrue()), nil
}
