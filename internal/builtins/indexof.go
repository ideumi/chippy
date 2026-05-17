/*
 *
 * RR2 - internal/builtins/indexof.go
 *
 */

package builtins

import (
	"bytes"
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"strings"
)

func indexofFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("indexof", 2, "haystack, needle")))
	}

	switch haystack := args[0].(type) {
	case *values.String:
		needleArg, ok := args[1].(*values.String)

		if !ok {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("indexof", shared.PositionSecond, shared.TypeString, "needle")))
		}

		needle := needleArg.Value

		if len(needle) == 0 {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		byteIndex := strings.Index(haystack.Value, needle)

		if byteIndex == -1 {
			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
		}

		runeIndex := len([]rune(haystack.Value[:byteIndex]))

		return res.Success(values.NewNumber(runeIndex + 1).SetContext(ctx))

	case *values.List:
		needle := args[1]

		for i, element := range haystack.Elements {
			comparison, err := element.GetComparisonEe(needle)

			if err != nil {
				continue
			}

			if compNum, ok := comparison.(*values.Number); ok {
				if compNum.IsTrue() {
					return res.Success(values.NewNumber(i + 1).SetContext(ctx))
				}
			}
		}

		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))

	case *values.Bytes:
		switch needle := args[1].(type) {
		case *values.Number:
			byte64, err := needle.AsInt()

			if err != nil {
				return res.Failure(err)
			}

			byteValue := int(byte64)

			if byteValue < 0 || byteValue > 255 {
				posStart, posEnd := args[1].GetPos()

				return res.Failure(errors.NewRTError(
					posStart, posEnd,
					"Byte values must be between 0 and 255"))
			}

			target := byte(byteValue)

			for i, b := range haystack.Data {
				if b == target {
					return res.Success(values.NewNumber(i + 1).SetContext(ctx))
				}
			}

			return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))

		case *values.Bytes:
			if len(needle.Data) == 0 {
				return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
			}

			idx := bytes.Index(haystack.Data, needle.Data)

			if idx == -1 {
				return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
			}

			return res.Success(values.NewNumber(idx + 1).SetContext(ctx))

		default:
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidArgTypePositionalWithHint("indexof", shared.PositionSecond, "a number or bytes", "needle")))
		}

	default:
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("indexof", shared.PositionFirst, "a string, list, or bytes", "haystack")))
	}
}
