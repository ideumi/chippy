/*
 *
 * RR2 - internal/builtins/has.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"strings"
)

func hasFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("has", 2, "container, needle"),
			ctx,
		))
	}

	switch container := args[0].(type) {
	case *values.Map:
		keyStr, ok := args[1].(*values.String)

		if !ok {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeString, "key"),
				ctx,
			))
		}

		if _, exists := container.Entries[keyStr.Value]; exists {
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
				if compNum.Value == constants.NUM_TRU {
					return res.Success(values.NewNumber(constants.NUM_TRU).SetContext(ctx))
				}
			}
		}

		return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))

	case *values.String:
		needleStr, ok := args[1].(*values.String)

		if !ok {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeString, "needle"),
				ctx,
			))
		}

		if needleStr.Value == "" {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Cannot search for empty string",
				ctx,
			))
		}

		if strings.Contains(container.Value, needleStr.Value) {
			return res.Success(values.NewNumber(constants.NUM_TRU).SetContext(ctx))
		}

		return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))

	case *values.Bytes:
		needleNum, ok := args[1].(*values.Number)

		if !ok {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionSecond, shared.TypeNumber, "bytes"),
				ctx,
			))
		}

		byteValue := int(needleNum.Value)

		if byteValue < 0 || byteValue > 255 {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Byte values must be between 0 and 255",
				ctx,
			))
		}

		target := byte(byteValue)

		for _, b := range container.Data {
			if b == target {
				return res.Success(values.NewNumber(constants.NUM_TRU).SetContext(ctx))
			}
		}

		return res.Success(values.NewNumber(constants.NUM_FAL).SetContext(ctx))

	default:
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("has", shared.PositionFirst, "a map, list, string, or bytes", "container"),
			ctx,
		))
	}
}
