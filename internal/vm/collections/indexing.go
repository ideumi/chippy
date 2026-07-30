/*
 *
 * Modena - internal/vm/collections/indexing.go
 *
 */

package collections

import (
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

// Part is which piece of x[i] = v a fault blames, so the vm can underline the
// 99 in l[99] rather than the whole expression. A working index never looks a
// position up at all.
type Part int

const (
	PartCollection Part = 0
	PartIndex      Part = 1
	PartValue      Part = 2
)

type Fault struct {
	Part Part
	Err  error
}

func (f *Fault) Error() string {
	return f.Err.Error()
}

func faultAt(part Part, message string) error {
	return &Fault{Part: part, Err: errors.NewCallError(message)}
}

func faultWrap(part Part, err error) error {
	return &Fault{Part: part, Err: err}
}

func Get(collection, index values.Value) (values.Value, error) {
	if mapValue, ok := values.AsMap(collection); ok {
		key, ok := values.AsString(index)

		if !ok {
			return values.Value{}, faultAt(PartIndex, "Map index must be a string")
		}

		if value, exists := mapValue.Entries[key.Value]; exists {
			return value, nil
		}

		return values.Value{}, faultAt(PartIndex, "Map key not found")
	}

	idx, err := intIndex(index)

	if err != nil {
		return values.Value{}, err
	}

	if list, ok := values.AsList(collection); ok {
		if idx < 1 || idx > len(list.Elements) {
			return values.Value{}, faultAt(PartIndex, "Index out of bounds")
		}

		return list.Elements[idx-1], nil
	}

	if bytesVal, ok := values.AsBytes(collection); ok {
		if idx < 1 || idx > len(bytesVal.Data) {
			return values.Value{}, faultAt(PartIndex, "Index out of bounds")
		}

		return values.NewNumber(bytesVal.Data[idx-1]), nil
	}

	if str, ok := values.AsString(collection); ok {
		if char, inRange := str.RuneAt(idx); inRange {
			return values.NewString(char), nil
		}

		return values.Value{}, faultAt(PartIndex, "Index out of bounds")
	}

	return values.Value{}, faultAt(PartCollection, "Can only index lists, maps, bytes, or strings")
}

func Set(collection, index, value values.Value) (values.Value, error) {
	if mapValue, ok := values.AsMap(collection); ok {
		key, ok := values.AsString(index)

		if !ok {
			return values.Value{}, faultAt(PartIndex, "Map index must be a string")
		}

		mapValue.Set(key.Value, value)

		return collection, nil
	}

	idx, err := intIndex(index)

	if err != nil {
		return values.Value{}, err
	}

	if list, ok := values.AsList(collection); ok {
		if idx < 1 || idx > len(list.Elements) {
			return values.Value{}, faultAt(PartIndex, "Index out of bounds")
		}

		list.Elements[idx-1] = value

		return collection, nil
	}

	if bytesVal, ok := values.AsBytes(collection); ok {
		if !value.IsNumber() {
			return values.Value{}, faultAt(PartValue, "Byte value must be a number")
		}

		byteVal, err := byteFromNumber(value)

		if err != nil {
			return values.Value{}, faultWrap(PartValue, err)
		}

		if idx < 1 || idx > len(bytesVal.Data) {
			return values.Value{}, faultAt(PartIndex, "Index out of bounds")
		}

		bytesVal.Data[idx-1] = byteVal

		return collection, nil
	}

	return values.Value{}, faultAt(PartCollection, "Can only assign to list, map, or bytes indices")
}

func intIndex(index values.Value) (int, error) {
	if !index.IsNumber() {
		return 0, faultAt(PartIndex, "Index must be a number")
	}

	if !index.IsInt() {
		return 0, faultAt(PartIndex, "Index must be an integer")
	}

	idx, err := index.AsInt()

	if err != nil {
		return 0, err
	}

	return int(idx), nil
}
