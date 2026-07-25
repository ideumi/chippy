/*
 *
 * RR2 - internal/builtins/has.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"strings"
)

func hasFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("has", 2, "container, needle"))
	}

	switch container := args[0].(type) {
	case *values.String:
		needleStr, ok := args[1].(*values.String)

		if !ok {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeString, "needle"))
		}

		if needleStr.Value == "" {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		if strings.Contains(container.Value, needleStr.Value) {
			return res.Success(values.NewNumber(constants.NUM_TRU).SetContext(ctx))
		}

		return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))

	case *values.List:
		needle := args[1]

		for _, element := range container.Elements {
			if element == nil {
				continue
			}

			comparison, err := element.GetComparisonEe(needle)

			if err != nil {
				continue
			}

			if compNum, ok := comparison.(*values.Number); ok {
				if compNum.IsTrue() {
					return res.Success(values.NewNumber(constants.NUM_TRU).SetContext(ctx))
				}
			}
		}

		return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))

	case *values.Bytes:
		needleNum, ok := args[1].(*values.Number)

		if !ok {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeNumber, "bytes"))
		}

		byte64, err := needleNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		byteValue := int(byte64)

		if byteValue < 0 || byteValue > 255 {
			return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))
		}

		target := byte(byteValue)

		for _, b := range container.Data {
			if b == target {
				return res.Success(values.NewNumber(constants.NUM_TRU).SetContext(ctx))
			}
		}

		return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))

	case *values.Map:
		keyStr, ok := args[1].(*values.String)

		if !ok {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeString, "key"))
		}

		if _, exists := container.Entries[keyStr.Value]; exists {
			return res.Success(values.NewNumber(constants.NUM_TRU).SetContext(ctx))
		}

		return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))

	default:
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionFirst, "a string, list, bytes, or map", "container"))
	}
}
