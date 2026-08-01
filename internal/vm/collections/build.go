/*
 *
 * Chippy - internal/vm/collections/build.go
 *
 */

package collections

import (
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

// Errors here are left unpositioned. The vm knows which instruction it is running
// and puts the position on before the error reaches the user.
func BuildBytes(elements []values.Value) ([]byte, error) {
	data := make([]byte, len(elements))

	for idx, element := range elements {
		if !element.IsNumber() {
			return nil, errors.NewCallError("Byte array elements must be numbers")
		}

		value, err := byteFromNumber(element)

		if err != nil {
			return nil, err
		}

		data[idx] = value
	}

	return data, nil
}

func byteFromNumber(num values.Value) (byte, error) {
	value, err := num.AsInt()

	if err != nil {
		return 0, err
	}

	if value < 0 || value > 255 {
		return 0, errors.NewCallError("Byte values must be between 0 and 255")
	}

	return byte(value), nil
}

func BuildMap(pairs []values.Value) (values.Value, error) {
	keys := make([]string, 0, len(pairs)/2)
	entries := make(map[string]values.Value, len(pairs)/2)

	for i := 0; i < len(pairs); i += 2 {
		key, ok := values.AsString(pairs[i])

		if !ok {
			return values.Value{}, errors.NewCallError("Map keys must be strings")
		}

		if _, exists := entries[key.Value]; !exists {
			keys = append(keys, key.Value)
		}

		entries[key.Value] = pairs[i+1]
	}

	return values.NewMapFromEntries(keys, entries), nil
}
