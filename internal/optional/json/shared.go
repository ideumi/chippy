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
)

const (
	mapMarker = "map"
	jsonTrue  = "__~JSONTRUE~__"
	jsonFalse = "__~JSONFALSE~__"
	jsonNull  = "__~JSONNULL~__"
)

// goToLibMap converts Go JSON values to ChipLang libmap format
func goToLibMap(val interface{}, ctx interface{}) values.Value {
	switch v := val.(type) {
	case nil:
		return values.NewString(jsonNull).SetContext(ctx)
	case bool:
		if v {
			return values.NewString(jsonTrue).SetContext(ctx)
		}

		return values.NewString(jsonFalse).SetContext(ctx)
	case float64:
		return values.NewNumber(v).SetContext(ctx)
	case string:
		return values.NewString(v).SetContext(ctx)
	case []interface{}:
		list := values.NewList([]values.Value{})

		for _, item := range v {
			list.Elements = append(list.Elements, goToLibMap(item, ctx))
		}

		return list.SetContext(ctx)
	case map[string]interface{}:
		// ["map", [key1, val1], [key2, val2], ...] (libmap.chh)
		list := values.NewList([]values.Value{values.NewString(mapMarker).SetContext(ctx)})

		for key, value := range v {
			pair := values.NewList([]values.Value{
				values.NewString(key).SetContext(ctx),
				goToLibMap(value, ctx),
			})

			list.Elements = append(list.Elements, pair.SetContext(ctx))
		}

		return list.SetContext(ctx)
	default:
		return values.NewString(constants.STR_ERR).SetContext(ctx)
	}
}

// libMapToGo converts ChipLang libmap format to Go JSON values
func libMapToGo(val values.Value) (interface{}, error) {
	switch v := val.(type) {
	case *values.Number:
		return v.Value, nil
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
	case *values.List:
		if len(v.Elements) == 0 {
			return []interface{}{}, nil
		}

		// Check if it's a map ["map", [key, val], ...]
		if firstElem, ok := v.Elements[0].(*values.String); ok {
			if firstElem.Value == mapMarker {
				// Convert to Go map
				result := make(map[string]interface{})

				for i := 1; i < len(v.Elements); i++ {
					pair, ok := v.Elements[i].(*values.List)

					if !ok || len(pair.Elements) != 2 {
						return nil, fmt.Errorf("invalid map pair")
					}

					key, ok := pair.Elements[0].(*values.String)

					if !ok {
						return nil, fmt.Errorf("map key must be string")
					}

					value, err := libMapToGo(pair.Elements[1])

					if err != nil {
						return nil, err
					}

					result[key.Value] = value
				}

				return result, nil
			}
		}

		// Regular list (JSON array)
		result := make([]interface{}, len(v.Elements))

		for i, elem := range v.Elements {
			goVal, err := libMapToGo(elem)

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
