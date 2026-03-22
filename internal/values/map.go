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

func (m *Map) String() string {
	pairs := make([]string, len(m.Keys))

	for i, key := range m.Keys {
		val := m.Entries[key]

		var valStr string

		if val != nil {
			valStr = val.String()
		} else {
			valStr = "null"
		}

		pairs[i] = "\"" + key + "\": " + valStr
	}
	return "m[" + strings.Join(pairs, ", ") + "]"
}

func (m *Map) SetPos(posStart, posEnd *errors.Position) Value {
	m.BaseValue.SetPos(posStart, posEnd)

	return m
}

func (m *Map) SetContext(context interface{}) Value {
	m.BaseValue.SetContext(context)

	return m
}

func (m *Map) Copy() Value {
	newKeys := make([]string, len(m.Keys))
	copy(newKeys, m.Keys)
	newEntries := make(map[string]Value, len(m.Entries))

	for k, v := range m.Entries {
		if v != nil {
			newEntries[k] = v.Copy()
		} else {
			newEntries[k] = nil
		}
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

func (m *Map) MapSet(key string, val Value) *Map {
	newMap := m.Copy().(*Map)

	if _, exists := newMap.Entries[key]; !exists {
		newMap.Keys = append(newMap.Keys, key)
	}

	newMap.Entries[key] = val

	return newMap
}

func (m *Map) MapRemove(key string) (*Map, bool) {
	if _, exists := m.Entries[key]; !exists {
		return nil, false
	}

	newMap := m.Copy().(*Map)
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

	result := m.Copy().(*Map)

	for _, key := range otherMap.Keys {
		result = result.MapSet(key, otherMap.Entries[key])
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

		comparison, err := val.GetComparisonEe(otherVal)

		if err != nil {
			return nil, err
		}

		if compNum, ok := comparison.(*Number); ok {
			if compNum.Value == constants.NUM_FAL {
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
		var result float64

		if compNum.Value == constants.NUM_TRU {
			result = constants.NUM_FAL
		} else {
			result = constants.NUM_TRU
		}

		return NewNumber(result).SetContext(m.context), nil
	}

	return nil, IllegalOperation(m, other)
}

func (m *Map) Notted() (Value, error) {
	var result float64

	if m.IsTrue() {
		result = constants.NUM_FAL
	} else {
		result = constants.NUM_TRU
	}

	return NewNumber(result).SetContext(m.context), nil
}
