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
	return l.stringWalk(map[Value]bool{}, 0)
}

func (l *List) stringWalk(seen map[Value]bool, depth int) string {
	if depth > constants.LIMIT_VALUE_NESTING_DEPTH || seen[l] {
		return "[...]"
	}

	seen[l] = true
	defer delete(seen, l)

	elements := make([]string, len(l.Elements))

	for i, element := range l.Elements {
		elements[i] = walkString(element, seen, depth+1)
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
	return l.copyWalk(map[Value]bool{}, 0)
}

func (l *List) copyWalk(seen map[Value]bool, depth int) Value {
	defer EnterWalk(l, seen, depth)()

	newElements := make([]Value, len(l.Elements))

	for i, element := range l.Elements {
		newElements[i] = walkCopy(element, seen, depth+1)
	}

	copy := NewList(newElements)
	copy.SetPos(l.posStart, l.posEnd)
	copy.SetContext(l.context)

	return copy
}

func (l *List) ShallowCopy() *List {
	newElements := make([]Value, len(l.Elements))
	copy(newElements, l.Elements)

	result := NewList(newElements)
	result.SetPos(l.posStart, l.posEnd)
	result.SetContext(l.context)

	return result
}

func (l *List) IsTrue() bool {
	return len(l.Elements) > 0
}

func (l *List) AddedTo(other Value) (Value, error) {
	if otherList, ok := other.(*List); ok {
		newList := l.ShallowCopy()
		newList.Elements = append(newList.Elements, otherList.Elements...)

		return newList, nil
	}

	return nil, IllegalOperation(l, other)
}

func (l *List) SubbedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		if !otherNum.IsInt() {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Index must be an integer")

		}

		newList := l.ShallowCopy()
		index := int(otherNum.iVal)

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
		if !otherNum.IsInt() {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Repeat count must be an integer")

		}

		if otherNum.iVal < 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Cannot repeat list negative times")

		}

		newElements := []Value{}
		for i := int64(0); i < otherNum.iVal; i++ {
			newElements = append(newElements, l.Elements...)
		}

		result := NewList(newElements)
		result.SetContext(l.context)

		return result, nil
	}

	return nil, IllegalOperation(l, other)
}

func (l *List) GetComparisonEe(other Value) (Value, error) {
	return l.eqWalk(other, map[Value]bool{}, 0)
}

func (l *List) eqWalk(other Value, seen map[Value]bool, depth int) (Value, error) {
	defer EnterWalk(l, seen, depth)()

	if otherList, ok := other.(*List); ok {
		result := constants.NUM_TRU

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

				comparison, err := walkEqual(element, otherElement, seen, depth+1)
				if err != nil {
					return nil, err
				}

				if compNum, ok := comparison.(*Number); ok {
					if !compNum.IsTrue() {
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
		result := constants.NUM_TRU

		if compNum.IsTrue() {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(l.context), nil
	}

	return nil, IllegalOperation(l, other)
}

func (l *List) Notted() (Value, error) {
	result := constants.NUM_TRU

	if l.IsTrue() {
		result = constants.NUM_FAL
	}

	return NewNumber(result).SetContext(l.context), nil
}

func (l *List) XoredBy(other Value) (Value, error) {
	result := constants.NUM_FAL

	if l.IsTrue() != other.IsTrue() {
		result = constants.NUM_TRU
	}

	return NewNumber(result).SetContext(l.context), nil
}
