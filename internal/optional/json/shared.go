/*
 *
 * RR2 - internal/optional/json/shared.go
 *
 */

package json

import (
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"fmt"
	"sort"
)

const (
	jsonTrue  = "__~JSONTRUE~__"
	jsonFalse = "__~JSONFALSE~__"
	jsonNull  = "__~JSONNULL~__"
)

// unmarshalValue converts Go JSON values to Chippy values
func unmarshalValue(val interface{}, ctx values.Ctx) values.Value {
	switch typed := val.(type) {
	case nil:
		return values.NewString(jsonNull)
	case bool:
		if typed {
			return values.NewString(jsonTrue)
		}

		return values.NewString(jsonFalse)
	case float64:
		num, err := values.NewNumberFromFloat(typed)

		if err != nil {
			return values.NewString(constants.STR_ERR)
		}

		return num
	case string:
		return values.NewString(typed)
	case []interface{}:
		elements := make([]values.Value, 0, len(typed))

		for _, item := range typed {
			elements = append(elements, unmarshalValue(item, ctx))
		}

		return values.NewList(elements)
	case map[string]interface{}:
		/* NOTE:
		 * Sort keys for repeatable output.
		 * Work around https://go.dev/doc/go1#iteration instead of silently
		 * breaking things...
		 */
		keys := make([]string, 0, len(typed))

		for key := range typed {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		entries := make(map[string]values.Value, len(typed))

		for _, key := range keys {
			entries[key] = unmarshalValue(typed[key], ctx)
		}

		return values.NewMapFromEntries(keys, entries)
	default:
		return values.NewString(constants.STR_ERR)
	}
}

// marshalValue converts Chippy values to Go JSON values
func marshalValue(val values.Value, seen map[values.Value]bool, depth int) (interface{}, error) {
	if val.IsNumber() {
		if val.IsInt() {
			intVal, err := val.AsInt()

			if err != nil {
				return nil, err
			}

			return intVal, nil
		}

		return val.AsFloat(), nil
	}

	if str, ok := values.AsString(val); ok {
		// Check for special JSON constants
		switch str.Value {
		case jsonTrue:
			return true, nil
		case jsonFalse:
			return false, nil
		case jsonNull:
			return nil, nil
		default:
			return str.Value, nil
		}
	}

	if mapVal, ok := values.AsMap(val); ok {
		defer values.EnterWalk(val, seen, depth)()

		result := make(map[string]interface{})

		for _, key := range mapVal.Keys {
			entry := mapVal.Entries[key]

			if entry.IsUnset() {
				result[key] = nil
				continue
			}

			goVal, err := marshalValue(entry, seen, depth+1)

			if err != nil {
				return nil, err
			}

			result[key] = goVal
		}

		return result, nil
	}

	if list, ok := values.AsList(val); ok {
		defer values.EnterWalk(val, seen, depth)()

		result := make([]interface{}, len(list.Elements))

		for i, elem := range list.Elements {
			goVal, err := marshalValue(elem, seen, depth+1)

			if err != nil {
				return nil, err
			}

			result[i] = goVal
		}

		return result, nil
	}

	return nil, fmt.Errorf("unsupported type")
}
