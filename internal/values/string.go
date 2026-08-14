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
	"unicode/utf8"
)

type String struct {
	OperatorDefaults
	Value string
	runes int
}

var (
	emptyString = newCountedString("", 0)
	singleASCII = buildSingleASCII()
)

func buildSingleASCII() [utf8.RuneSelf]Value {
	var table [utf8.RuneSelf]Value

	for code := range table {
		table[code] = newCountedString(string(rune(code)), 1)
	}

	return table
}

func NewString(value string) Value {
	return newCountedString(value, utf8.RuneCountInString(value))
}

func newCountedString(value string, runes int) Value {
	return fromHeap(TagText, &String{Value: value, runes: runes})
}

func (s *String) RuneCount() int {
	return s.runes
}

func (s *String) IsASCII() bool {
	return s.runes == len(s.Value)
}

func (s *String) RuneAt(index int) (Value, bool) {
	if index < 1 || index > s.runes {
		return Value{}, false
	}

	if s.IsASCII() {
		return singleASCII[s.Value[index-1]], true
	}

	remaining := index

	for _, char := range s.Value {
		remaining--

		if remaining == 0 {
			return newCountedString(string(char), 1), true
		}
	}

	return Value{}, false
}

func (s *String) RuneSlice(start, end int) Value {
	if start < 1 || end > s.runes || end < start {
		return emptyString
	}

	count := end - start + 1

	if s.IsASCII() {
		return newCountedString(s.Value[start-1:end], count)
	}

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
		return emptyString
	}

	return newCountedString(s.Value[from:to], count)
}

func (s *String) Runes() []Value {
	elements := make([]Value, 0, s.runes)

	if s.IsASCII() {
		for index := 0; index < len(s.Value); index++ {
			elements = append(elements, singleASCII[s.Value[index]])
		}

		return elements
	}

	for _, char := range s.Value {
		elements = append(elements, newCountedString(string(char), 1))
	}

	return elements
}

func (s *String) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

// Shares the backing object. Nothing may mutate a String after construction.
func (s *String) Copy() Value {
	return fromHeap(TagText, s)
}

func (s *String) IsTrue() bool {
	return len(s.Value) > 0
}

func (s *String) AddedTo(other Value) (Value, error) {
	otherStr, ok := AsString(other)

	if !ok {
		return Value{}, IllegalOperation()
	}

	if len(s.Value) == 0 {
		return other, nil
	}

	if len(otherStr.Value) == 0 {
		return fromHeap(TagText, s), nil
	}

	return newCountedString(s.Value+otherStr.Value, s.runes+otherStr.runes), nil
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

	return newCountedString(strings.Repeat(s.Value, int(count)), s.runes*int(count)), nil
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
