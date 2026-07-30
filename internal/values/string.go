/*
 *
 * Chippy - internal/values/string.go
 *
 */

package values

import (
	"chip-go/internal/errors"
	"fmt"
	"strings"
)

type String struct {
	OperatorDefaults
	Value string
}

func NewString(value string) Value {
	return fromHeap(TagText, &String{Value: value})
}

func (s *String) RuneAt(index int) (string, bool) {
	if index < 1 {
		return "", false
	}

	for _, char := range s.Value {
		index--

		if index == 0 {
			return string(char), true
		}
	}

	return "", false
}

func (s *String) RuneSlice(start, end int) string {
	from, to := -1, len(s.Value)
	index := 0

	for offset := range s.Value {
		index++

		if index == start {
			from = offset
		}

		if index == end+1 {
			to = offset

			break
		}
	}

	if from < 0 {
		return ""
	}

	return s.Value[from:to]
}

func (s *String) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

func (s *String) Copy() Value {
	return NewString(s.Value)
}

func (s *String) IsTrue() bool {
	return len(s.Value) > 0
}

func (s *String) AddedTo(other Value) (Value, error) {
	if otherStr, ok := AsString(other); ok {
		return NewString(s.Value + otherStr.Value), nil
	}

	return Value{}, IllegalOperation()
}

func (s *String) MultedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if !other.IsInt() {
		return Value{}, errors.NewCallError("Repeat count must be an integer")
	}

	count, _ := other.AsInt()

	if count < 0 {
		return Value{}, errors.NewCallError("Cannot repeat string negative times")
	}

	return NewString(strings.Repeat(s.Value, int(count))), nil
}

func (s *String) GetComparisonEe(other Value) (Value, error) {
	if otherStr, ok := AsString(other); ok {
		return Bool(s.Value == otherStr.Value), nil
	}

	return Value{}, IllegalOperation()
}

func (s *String) GetComparisonNe(other Value) (Value, error) {
	if otherStr, ok := AsString(other); ok {
		return Bool(s.Value != otherStr.Value), nil
	}

	return Value{}, IllegalOperation()
}

func (s *String) Notted() (Value, error) {
	return Bool(!s.IsTrue()), nil
}

func (s *String) XoredBy(other Value) (Value, error) {
	return Bool(s.IsTrue() != other.IsTrue()), nil
}
