/*
 *
 * Chippy - internal/builtins/indexof.go
 *
 */

package builtins

import (
	"bytes"
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"strings"
	"unicode/utf8"
)

func indexofFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("indexof", 2, "haystack, needle"))
	}

	if container, ok := values.AsString(args[0]); ok {
		needleArg, ok := values.AsString(args[1])

		if !ok {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("indexof", shared.PositionSecond, shared.TypeString, "needle"))
		}

		needle := needleArg.Value

		if len(needle) == 0 {
			return res.Success(values.NewString(constants.STR_ERR))
		}

		byteIndex := strings.Index(container.Value, needle)

		if byteIndex == -1 {
			return res.Success(values.NewString(constants.STR_ERR))
		}

		if container.IsASCII() {
			return res.Success(values.NewNumber(byteIndex + 1))
		}

		runeIndex := utf8.RuneCountInString(container.Value[:byteIndex])

		return res.Success(values.NewNumber(runeIndex + 1))
	}

	if container, ok := values.AsList(args[0]); ok {
		needle := args[1]

		for index, element := range container.Elements {
			comparison, err := element.GetComparisonEe(needle)

			if err != nil {
				continue
			}

			if comparison.IsNumber() && comparison.IsTrue() {
				return res.Success(values.NewNumber(index + 1))
			}
		}

		return res.Success(values.NewString(constants.STR_ERR))
	}

	if container, ok := values.AsBytes(args[0]); ok {
		if needle := args[1]; needle.IsNumber() {
			byte64, err := needle.AsInt()

			if err != nil {
				return res.Failure(err)
			}

			byteValue := int(byte64)

			if byteValue < 0 || byteValue > 255 {
				return res.Success(values.NewString(constants.STR_ERR))
			}

			target := byte(byteValue)

			for index, byteVal := range container.Data {
				if byteVal == target {
					return res.Success(values.NewNumber(index + 1))
				}
			}

			return res.Success(values.NewString(constants.STR_ERR))
		}

		if needle, ok := values.AsBytes(args[1]); ok {
			if len(needle.Data) == 0 {
				return res.Success(values.NewString(constants.STR_ERR))
			}

			idx := bytes.Index(container.Data, needle.Data)

			if idx == -1 {
				return res.Success(values.NewString(constants.STR_ERR))
			}

			return res.Success(values.NewNumber(idx + 1))
		}

		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("indexof", shared.PositionSecond, "a number or bytes", "needle"))
	}

	return res.FailAt(1,
		shared.Errors.InvalidArgTypePositionalWithHint("indexof", shared.PositionFirst, "a string, list, or bytes", "haystack"))
}
