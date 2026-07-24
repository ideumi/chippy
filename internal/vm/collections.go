/*
 *
 * Modena - internal/vm/collections.go
 *
 */

package vm

import (
	"chip-go/internal/bytecode"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func bytesFromStack(elements []values.Value, span bytecode.Span) ([]byte, error) {
	data := make([]byte, len(elements))

	for idx, element := range elements {
		num, ok := element.(*values.Number)

		if !ok {
			return nil, errors.NewRTError(span.Start, span.End, "Byte array elements must be numbers")
		}

		value, err := byteFromNumber(num, span)

		if err != nil {
			return nil, err
		}

		data[idx] = value
	}

	return data, nil
}

func byteFromNumber(num *values.Number, span bytecode.Span) (byte, error) {
	value, err := num.AsInt()

	if err != nil {
		return 0, err
	}

	if value < 0 || value > 255 {
		return 0, errors.NewRTError(span.Start, span.End, "Byte values must be between 0 and 255")
	}

	return byte(value), nil
}

func mapFromStack(pairs []values.Value, ctx values.Ctx) (values.Value, error) {
	keys := make([]string, 0, len(pairs)/2)
	entries := make(map[string]values.Value, len(pairs)/2)

	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(*values.String)

		if !ok {
			posStart, posEnd := pairs[i].GetPos()
			return nil, errors.NewRTError(posStart, posEnd, "Map keys must be strings")
		}

		if _, exists := entries[key.Value]; !exists {
			keys = append(keys, key.Value)
		}

		entries[key.Value] = pairs[i+1].SetContext(ctx)
	}

	return values.NewMapFromEntries(keys, entries).SetContext(ctx), nil
}

func indexGet(collection, index values.Value, ctx values.Ctx, collSpan, idxSpan bytecode.Span) (values.Value, error) {
	if mapValue, ok := collection.(*values.Map); ok {
		key, ok := index.(*values.String)

		if !ok {
			return nil, errors.NewRTError(idxSpan.Start, idxSpan.End, "Map index must be a string")
		}

		if value, exists := mapValue.Entries[key.Value]; exists {
			return value, nil
		}

		return values.NewString(constants.STR_ERR).SetContext(ctx), nil
	}

	idx, err := intIndex(index, idxSpan)

	if err != nil {
		return nil, err
	}

	switch coll := collection.(type) {
	case *values.List:
		if idx < 1 || idx > len(coll.Elements) {
			return nil, errors.NewRTError(idxSpan.Start, idxSpan.End, "Index out of bounds")
		}

		return coll.Elements[idx-1], nil

	case *values.Bytes:
		if idx < 1 || idx > len(coll.Data) {
			return nil, errors.NewRTError(idxSpan.Start, idxSpan.End, "Index out of bounds")
		}

		return values.NewNumber(coll.Data[idx-1]).SetContext(ctx), nil

	case *values.String:
		runes := []rune(coll.Value)

		if idx < 1 || idx > len(runes) {
			return nil, errors.NewRTError(idxSpan.Start, idxSpan.End, "Index out of bounds")
		}

		return values.NewString(string(runes[idx-1])).SetContext(ctx), nil

	default:
		return nil, errors.NewRTError(collSpan.Start, collSpan.End, "Can only index lists, maps, bytes, or strings")
	}
}

func indexSet(collection, index, value values.Value, ctx values.Ctx, collSpan, idxSpan, valSpan bytecode.Span) (values.Value, error) {
	if mapValue, ok := collection.(*values.Map); ok {
		key, ok := index.(*values.String)

		if !ok {
			return nil, errors.NewRTError(idxSpan.Start, idxSpan.End, "Map index must be a string")
		}

		mapValue.Set(key.Value, value.SetContext(ctx))

		return mapValue, nil
	}

	idx, err := intIndex(index, idxSpan)

	if err != nil {
		return nil, err
	}

	switch coll := collection.(type) {
	case *values.List:
		if idx < 1 || idx > len(coll.Elements) {
			return nil, errors.NewRTError(idxSpan.Start, idxSpan.End, "Index out of bounds")
		}

		coll.Elements[idx-1] = value.SetContext(ctx)

		return coll, nil

	case *values.Bytes:
		num, ok := value.(*values.Number)

		if !ok {
			return nil, errors.NewRTError(valSpan.Start, valSpan.End, "Byte value must be a number")
		}

		byteVal, err := byteFromNumber(num, valSpan)

		if err != nil {
			return nil, err
		}

		if idx < 1 || idx > len(coll.Data) {
			return nil, errors.NewRTError(idxSpan.Start, idxSpan.End, "Index out of bounds")
		}

		coll.Data[idx-1] = byteVal

		return coll.SetContext(ctx), nil

	default:
		return nil, errors.NewRTError(collSpan.Start, collSpan.End, "Can only assign to list, map, or bytes indices")
	}
}

func intIndex(index values.Value, idxSpan bytecode.Span) (int, error) {
	num, ok := index.(*values.Number)

	if !ok {
		return 0, errors.NewRTError(idxSpan.Start, idxSpan.End, "Index must be a number")
	}

	if !num.IsInt() {
		return 0, errors.NewRTError(idxSpan.Start, idxSpan.End, "Index must be an integer")
	}

	idx, err := num.AsInt()

	if err != nil {
		return 0, err
	}

	return int(idx), nil
}
