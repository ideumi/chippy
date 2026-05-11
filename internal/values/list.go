/*
 *
 * RR2 - internal/values/list.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"strings"
)

type List struct {
	*BaseValue
	Elements []Value
}

func NewList(elements []Value) *List {
	return &List{
		BaseValue: NewBaseValue(),
		Elements:  elements,
	}
}

func (l *List) String() string {
	elements := make([]string, len(l.Elements))
	for i, element := range l.Elements {
		if element != nil {
			elements[i] = element.String()
		} else {
			elements[i] = "null"
		}
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

func (l *List) SetPos(posStart, posEnd *errors.Position) Value {
	l.BaseValue.SetPos(posStart, posEnd)

	return l
}

func (l *List) SetContext(ctx Ctx) Value {
	l.BaseValue.SetContext(ctx)

	return l
}

func (l *List) Copy() Value {
	newElements := make([]Value, len(l.Elements))
	for i, element := range l.Elements {
		if element != nil {
			newElements[i] = element.Copy()
		}
	}
	copy := NewList(newElements)
	copy.SetPos(l.posStart, l.posEnd)
	copy.SetContext(l.context)

	return copy
}

func (l *List) IsTrue() bool {
	return len(l.Elements) > 0
}

func (l *List) AddedTo(other Value) (Value, error) {
	if otherList, ok := other.(*List); ok {
		newList := l.Copy().(*List)
		newList.Elements = append(newList.Elements, otherList.Elements...)

		return newList, nil
	}

	return nil, IllegalOperation(l, other)
}

func (l *List) SubbedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		newList := l.Copy().(*List)
		index := int(otherNum.Value)

		if index < 1 || index > len(newList.Elements) {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Element at this index could not be removed from list because index is out of bounds")

		}

		newList.Elements = append(newList.Elements[:index-1], newList.Elements[index:]...)

		return newList, nil
	}

	return nil, IllegalOperation(l, other)
}

func (l *List) MultedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		if otherNum.Value < 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Cannot repeat list negative times")

		}

		newElements := []Value{}
		for i := 0; i < int(otherNum.Value); i++ {
			for _, element := range l.Elements {
				if element != nil {
					newElements = append(newElements, element.Copy())
				} else {
					newElements = append(newElements, nil)
				}
			}
		}

		result := NewList(newElements)
		result.SetContext(l.context)

		return result, nil
	}

	return nil, IllegalOperation(l, other)
}

func (l *List) GetComparisonEe(other Value) (Value, error) {
	if otherList, ok := other.(*List); ok {
		var result float64 = constants.NUM_TRU

		if len(l.Elements) != len(otherList.Elements) {
			result = constants.NUM_FAL
		} else {
			for i, element := range l.Elements {
				otherElement := otherList.Elements[i]

				if element == nil && otherElement == nil {
					continue
				}

				if element == nil || otherElement == nil {
					result = constants.NUM_FAL
					break
				}

				comparison, err := element.GetComparisonEe(otherElement)
				if err != nil {
					return nil, err
				}

				if compNum, ok := comparison.(*Number); ok {
					if compNum.Value == constants.NUM_FAL {
						result = constants.NUM_FAL
						break
					}
				}
			}
		}

		return NewNumber(result).SetContext(l.context), nil
	}

	return nil, IllegalOperation(l, other)
}

func (l *List) GetComparisonNe(other Value) (Value, error) {
	comparison, err := l.GetComparisonEe(other)

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

		return NewNumber(result).SetContext(l.context), nil
	}

	return nil, IllegalOperation(l, other)
}

func (l *List) Notted() (Value, error) {
	var result float64

	if l.IsTrue() {
		result = constants.NUM_FAL
	} else {
		result = constants.NUM_TRU
	}

	return NewNumber(result).SetContext(l.context), nil
}

func (l *List) XoredBy(other Value) (Value, error) {
	var result float64

	if l.IsTrue() != other.IsTrue() {
		result = constants.NUM_TRU
	} else {
		result = constants.NUM_FAL
	}

	return NewNumber(result).SetContext(l.context), nil
}
