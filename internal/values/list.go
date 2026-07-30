/*
 *
 * Modena - internal/values/list.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"strings"
)

type List struct {
	OperatorDefaults
	Elements []Value
}

func newListObject(elements []Value) *List {
	return &List{Elements: elements}
}

func NewList(elements []Value) Value {
	return fromHeap(TagList, newListObject(elements))
}

func (l *List) wrap() Value {
	return fromHeap(TagList, l)
}

func (l *List) String() string {
	return l.stringWalk(map[Value]bool{}, 0)
}

func (l *List) stringWalk(seen map[Value]bool, depth int) string {
	self := l.wrap()

	if depth > constants.LIMIT_VALUE_NESTING_DEPTH || seen[self] {
		return "[...]"
	}

	seen[self] = true
	defer delete(seen, self)

	elements := make([]string, len(l.Elements))

	for i, element := range l.Elements {
		elements[i] = walkString(element, seen, depth+1)
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

func (l *List) Copy() Value {
	return l.copyWalk(map[Value]bool{}, 0)
}

func (l *List) copyWalk(seen map[Value]bool, depth int) Value {
	defer EnterWalk(l.wrap(), seen, depth)()

	newElements := make([]Value, len(l.Elements))

	for i, element := range l.Elements {
		newElements[i] = walkCopy(element, seen, depth+1)
	}

	return NewList(newElements)
}

func (l *List) shallowCopy() *List {
	newElements := make([]Value, len(l.Elements))
	copy(newElements, l.Elements)

	return newListObject(newElements)
}

func (l *List) IsTrue() bool {
	return len(l.Elements) > 0
}

func (l *List) AddedTo(other Value) (Value, error) {
	if otherList, ok := AsList(other); ok {
		newList := l.shallowCopy()
		newList.Elements = append(newList.Elements, otherList.Elements...)

		return newList.wrap(), nil
	}

	return Value{}, IllegalOperation()
}

func (l *List) SubbedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if !other.IsInt() {
		return Value{}, errors.NewCallError("Index must be an integer")
	}

	newList := l.shallowCopy()
	position, _ := other.AsInt()
	index := int(position)

	if index < 1 || index > len(newList.Elements) {
		return Value{}, errors.NewCallError(
			"Element at this index could not be removed from list because index is out of bounds")
	}

	newList.Elements = append(newList.Elements[:index-1], newList.Elements[index:]...)

	return newList.wrap(), nil
}

func (l *List) MultedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if !other.IsInt() {
		return Value{}, errors.NewCallError("Repeat count must be an integer")
	}

	count, _ := other.AsInt()

	if count < 0 {
		return Value{}, errors.NewCallError("Cannot repeat list negative times")
	}

	newElements := []Value{}

	for i := int64(0); i < count; i++ {
		newElements = append(newElements, l.Elements...)
	}

	return NewList(newElements), nil
}

func (l *List) GetComparisonEe(other Value) (Value, error) {
	return l.eqWalk(other, map[Value]bool{}, 0)
}

func (l *List) eqWalk(other Value, seen map[Value]bool, depth int) (Value, error) {
	leave, err := enterWalk(l.wrap(), seen, depth)

	if err != nil {
		return Value{}, err
	}

	defer leave()

	otherList, ok := AsList(other)

	if !ok {
		return Value{}, IllegalOperation()
	}

	if len(l.Elements) != len(otherList.Elements) {
		return Bool(false), nil
	}

	for i, element := range l.Elements {
		otherElement := otherList.Elements[i]

		if element.IsUnset() && otherElement.IsUnset() {
			continue
		}

		if element.IsUnset() || otherElement.IsUnset() {
			return Bool(false), nil
		}

		comparison, err := walkEqual(element, otherElement, seen, depth+1)

		if err != nil {
			return Value{}, err
		}

		if !comparison.IsTrue() {
			return Bool(false), nil
		}
	}

	return Bool(true), nil
}

func (l *List) GetComparisonNe(other Value) (Value, error) {
	comparison, err := l.GetComparisonEe(other)

	if err != nil {
		return Value{}, err
	}

	return Bool(!comparison.IsTrue()), nil
}

func (l *List) Notted() (Value, error) {
	return Bool(!l.IsTrue()), nil
}

func (l *List) XoredBy(other Value) (Value, error) {
	return Bool(l.IsTrue() != other.IsTrue()), nil
}
