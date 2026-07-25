/*
 *
 * RR2 - internal/values/string.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"fmt"
	"strings"
)

type String struct {
	*BaseValue
	Value string
}

func NewString(value string) *String {
	return &String{
		BaseValue: NewBaseValue(),
		Value:     value,
	}
}

func (s *String) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

func (s *String) SetPos(posStart, posEnd *errors.Position) Value {
	s.BaseValue.SetPos(posStart, posEnd)
	return s
}

func (s *String) SetContext(ctx Ctx) Value {
	s.BaseValue.SetContext(ctx)
	return s
}

func (s *String) Copy() Value {
	copy := NewString(s.Value)
	copy.SetPos(s.posStart, s.posEnd)
	copy.SetContext(s.context)

	return copy
}

func (s *String) IsTrue() bool {
	return len(s.Value) > 0
}

func (s *String) AddedTo(other Value) (Value, error) {
	if otherStr, ok := other.(*String); ok {
		result := NewString(s.Value + otherStr.Value)
		result.SetContext(s.context)

		return result, nil
	}

	return nil, IllegalOperation()
}

func (s *String) MultedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		if !otherNum.IsInt() {
			return nil, errors.NewCallError(
				"Repeat count must be an integer")

		}

		if otherNum.iVal < 0 {
			return nil, errors.NewCallError(
				"Cannot repeat string negative times")

		}

		result := NewString(strings.Repeat(s.Value, int(otherNum.iVal)))
		result.SetContext(s.context)

		return result, nil
	}

	return nil, IllegalOperation()
}

func (s *String) GetComparisonEe(other Value) (Value, error) {
	if otherStr, ok := other.(*String); ok {
		result := constants.NUM_FAL

		if s.Value == otherStr.Value {
			result = constants.NUM_TRU
		}

		return NewNumber(result).SetContext(s.context), nil
	}

	return nil, IllegalOperation()
}

func (s *String) GetComparisonNe(other Value) (Value, error) {
	if otherStr, ok := other.(*String); ok {
		result := constants.NUM_FAL

		if s.Value != otherStr.Value {
			result = constants.NUM_TRU
		}

		return NewNumber(result).SetContext(s.context), nil
	}

	return nil, IllegalOperation()
}

func (s *String) Notted() (Value, error) {
	result := constants.NUM_TRU

	if s.IsTrue() {
		result = constants.NUM_FAL
	}

	return NewNumber(result).SetContext(s.context), nil
}

func (s *String) XoredBy(other Value) (Value, error) {
	result := constants.NUM_FAL

	if s.IsTrue() != other.IsTrue() {
		result = constants.NUM_TRU
	}

	return NewNumber(result).SetContext(s.context), nil
}
