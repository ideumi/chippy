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
	switch v := val.(type) {
	case nil:
		return values.NewString(jsonNull).SetContext(ctx)
	case bool:
		if v {
			return values.NewString(jsonTrue).SetContext(ctx)
		}

		return values.NewString(jsonFalse).SetContext(ctx)
	case float64:
		num, err := values.NewNumberFromFloat(v)

		if err != nil {
			return values.NewString(constants.STR_ERR).SetContext(ctx)
		}

		return num.SetContext(ctx)
	case string:
		return values.NewString(v).SetContext(ctx)
	case []interface{}:
		list := values.NewList([]values.Value{})

		for _, item := range v {
			list.Elements = append(list.Elements, unmarshalValue(item, ctx))
		}

		return list.SetContext(ctx)
	case map[string]interface{}:
		/* NOTE:
		 * Sort keys for repeatable output.
		 * Work around https://go.dev/doc/go1#iteration instead of silently
		 * breaking things...
		 */
		keys := make([]string, 0, len(v))

		for key := range v {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		entries := make(map[string]values.Value, len(v))

		for _, key := range keys {
			entries[key] = unmarshalValue(v[key], ctx)
		}

		return values.NewMapFromEntries(keys, entries).SetContext(ctx)
	default:
		return values.NewString(constants.STR_ERR).SetContext(ctx)
	}
}

// marshalValue converts Chippy values to Go JSON values
func marshalValue(val values.Value, seen map[values.Value]bool, depth int) (interface{}, error) {
	switch v := val.(type) {
	case *values.Number:
		if v.IsInt() {
			i, err := v.AsInt()

			if err != nil {
				return nil, err
			}

			return i, nil
		}

		return v.AsFloat(), nil
	case *values.String:
		// Check for special JSON constants
		switch v.Value {
		case jsonTrue:
			return true, nil
		case jsonFalse:
			return false, nil
		case jsonNull:
			return nil, nil
		default:
			return v.Value, nil
		}
	case *values.Map:
		defer values.EnterWalk(v, seen, depth)()

		result := make(map[string]interface{})

		for _, key := range v.Keys {
			val := v.Entries[key]

			if val == nil {
				result[key] = nil
				continue
			}

			goVal, err := marshalValue(val, seen, depth+1)

			if err != nil {
				return nil, err
			}

			result[key] = goVal
		}

		return result, nil
	case *values.List:
		defer values.EnterWalk(v, seen, depth)()

		result := make([]interface{}, len(v.Elements))

		for i, elem := range v.Elements {
			goVal, err := marshalValue(elem, seen, depth+1)

			if err != nil {
				return nil, err
			}

			result[i] = goVal
		}

		return result, nil
	default:
		return nil, fmt.Errorf("unsupported type")
	}
}
