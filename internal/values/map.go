/*
 *
 * RR2 - internal/values/map.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"strings"
)

type Map struct {
	*BaseValue
	Keys    []string
	Entries map[string]Value
}

func NewMap() *Map {
	return &Map{
		BaseValue: NewBaseValue(),
		Keys:      []string{},
		Entries:   map[string]Value{},
	}
}

func NewMapFromEntries(keys []string, entries map[string]Value) *Map {
	return &Map{
		BaseValue: NewBaseValue(),
		Keys:      keys,
		Entries:   entries,
	}
}

func (m *Map) String() string {
	return m.stringWalk(map[Value]bool{}, 0)
}

func (m *Map) stringWalk(seen map[Value]bool, depth int) string {
	if depth > constants.MAX_VALUE_DEPTH || seen[m] {
		return "m[...]"
	}

	seen[m] = true
	defer delete(seen, m)

	pairs := make([]string, len(m.Keys))

	for i, key := range m.Keys {
		valStr := walkString(m.Entries[key], seen, depth+1)

		pairs[i] = "\"" + key + "\": " + valStr
	}

	return "m[" + strings.Join(pairs, ", ") + "]"
}

func (m *Map) SetPos(posStart, posEnd *errors.Position) Value {
	m.BaseValue.SetPos(posStart, posEnd)

	return m
}

func (m *Map) SetContext(ctx Ctx) Value {
	m.BaseValue.SetContext(ctx)

	return m
}

func (m *Map) Copy() Value {
	return m.copyWalk(map[Value]bool{}, 0)
}

func (m *Map) copyWalk(seen map[Value]bool, depth int) Value {
	defer EnterWalk(m, seen, depth)()

	newKeys := make([]string, len(m.Keys))
	copy(newKeys, m.Keys)
	newEntries := make(map[string]Value, len(m.Entries))

	for k, v := range m.Entries {
		newEntries[k] = walkCopy(v, seen, depth+1)
	}

	result := &Map{
		BaseValue: NewBaseValue(),
		Keys:      newKeys,
		Entries:   newEntries,
	}

	result.SetPos(m.posStart, m.posEnd)
	result.SetContext(m.context)

	return result
}

func (m *Map) ShallowCopy() *Map {
	newKeys := make([]string, len(m.Keys))
	copy(newKeys, m.Keys)

	newEntries := make(map[string]Value, len(m.Entries))

	for k, v := range m.Entries {
		newEntries[k] = v
	}

	result := &Map{
		BaseValue: NewBaseValue(),
		Keys:      newKeys,
		Entries:   newEntries,
	}

	result.SetPos(m.posStart, m.posEnd)
	result.SetContext(m.context)

	return result
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

func (m *Map) MapRemove(key string) (*Map, bool) {
	if _, exists := m.Entries[key]; !exists {
		return nil, false
	}

	newMap := m.ShallowCopy()
	delete(newMap.Entries, key)
	newKeys := make([]string, 0, len(newMap.Keys)-1)

	for _, k := range newMap.Keys {
		if k != key {
			newKeys = append(newKeys, k)
		}
	}

	newMap.Keys = newKeys

	return newMap, true
}

func (m *Map) AddedTo(other Value) (Value, error) {
	otherMap, ok := other.(*Map)

	if !ok {
		return nil, IllegalOperation(m, other)
	}

	result := m.ShallowCopy()

	for _, key := range otherMap.Keys {
		result.Set(key, otherMap.Entries[key])
	}

	return result, nil
}

func (m *Map) SubbedBy(other Value) (Value, error) {
	otherStr, ok := other.(*String)

	if !ok {
		return nil, IllegalOperation(m, other)
	}

	newMap, found := m.MapRemove(otherStr.Value)

	if !found {
		return NewString(constants.STR_ERR).SetContext(m.context), nil
	}

	return newMap, nil
}

func (m *Map) GetComparisonEe(other Value) (Value, error) {
	return m.eqWalk(other, map[Value]bool{}, 0)
}

func (m *Map) eqWalk(other Value, seen map[Value]bool, depth int) (Value, error) {
	defer EnterWalk(m, seen, depth)()

	otherMap, ok := other.(*Map)

	if !ok {
		return NewNumber(constants.NUM_FAL).SetContext(m.context), nil
	}

	if len(m.Keys) != len(otherMap.Keys) {
		return NewNumber(constants.NUM_FAL).SetContext(m.context), nil
	}

	for key, val := range m.Entries {
		otherVal, exists := otherMap.Entries[key]

		if !exists {
			return NewNumber(constants.NUM_FAL).SetContext(m.context), nil
		}

		if val == nil && otherVal == nil {
			continue
		}

		if val == nil || otherVal == nil {
			return NewNumber(constants.NUM_FAL).SetContext(m.context), nil
		}

		comparison, err := walkEqual(val, otherVal, seen, depth+1)

		if err != nil {
			return nil, err
		}

		if compNum, ok := comparison.(*Number); ok {
			if !compNum.IsTrue() {
				return NewNumber(constants.NUM_FAL).SetContext(m.context), nil
			}
		}
	}

	return NewNumber(constants.NUM_TRU).SetContext(m.context), nil
}

func (m *Map) GetComparisonNe(other Value) (Value, error) {
	comparison, err := m.GetComparisonEe(other)

	if err != nil {
		return nil, err
	}

	if compNum, ok := comparison.(*Number); ok {
		result := constants.NUM_TRU

		if compNum.IsTrue() {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(m.context), nil
	}

	return nil, IllegalOperation(m, other)
}

func (m *Map) Notted() (Value, error) {
	result := constants.NUM_TRU

	if m.IsTrue() {
		result = constants.NUM_FAL
	}

	return NewNumber(result).SetContext(m.context), nil
}

func (m *Map) XoredBy(other Value) (Value, error) {
	result := constants.NUM_FAL

	if m.IsTrue() != other.IsTrue() {
		result = constants.NUM_TRU
	}

	return NewNumber(result).SetContext(m.context), nil
}
