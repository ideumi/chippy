/*
 *
 * Chippy - internal/builtins/has.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"strings"
)

func hasFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("has", 2, "container, needle"))
	}

	if container, ok := values.AsString(args[0]); ok {
		needleStr, ok := values.AsString(args[1])

		if !ok {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeString, "needle"))
		}

		if needleStr.Value == "" {
			return res.Success(values.NewString(constants.STR_ERR))
		}

		if strings.Contains(container.Value, needleStr.Value) {
			return res.Success(values.NewNumber(constants.NUM_TRU))
		}

		return res.Success(values.NewNumber(constants.NUM_FAL))
	}

	if container, ok := values.AsList(args[0]); ok {
		needle := args[1]

		for _, element := range container.Elements {
			if element.IsUnset() {
				continue
			}

			comparison, err := element.GetComparisonEe(needle)

			if err != nil {
				continue
			}

			if comparison.IsNumber() && comparison.IsTrue() {
				return res.Success(values.NewNumber(constants.NUM_TRU))
			}
		}

		return res.Success(values.NewNumber(constants.NUM_FAL))
	}

	if container, ok := values.AsBytes(args[0]); ok {
		needleNum := args[1]

		if !needleNum.IsNumber() {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeNumber, "bytes"))
		}

		byte64, err := needleNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		byteValue := int(byte64)

		if byteValue < 0 || byteValue > 255 {
			return res.Success(values.NewNumber(constants.NUM_FAL))
		}

		target := byte(byteValue)

		for _, byteVal := range container.Data {
			if byteVal == target {
				return res.Success(values.NewNumber(constants.NUM_TRU))
			}
		}

		return res.Success(values.NewNumber(constants.NUM_FAL))
	}

	if container, ok := values.AsMap(args[0]); ok {
		keyStr, ok := values.AsString(args[1])

		if !ok {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeString, "key"))
		}

		if _, exists := container.Entries[keyStr.Value]; exists {
			return res.Success(values.NewNumber(constants.NUM_TRU))
		}

		return res.Success(values.NewNumber(constants.NUM_FAL))
	}

	return res.FailAt(1,
		shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionFirst, "a string, list, bytes, or map", "container"))
}
