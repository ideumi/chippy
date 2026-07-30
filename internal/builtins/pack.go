/*
 *
 * RR2 - internal/builtins/pack.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func packFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("pack", 1, "value"))
	}

	value := args[0]

	if str, ok := values.AsString(value); ok {
		return res.Success(values.NewBytes([]byte(str.Value)))
	}

	if list, ok := values.AsList(value); ok {
		bytes := make([]byte, len(list.Elements))

		for i, element := range list.Elements {
			if element.IsUnset() {
				bytes[i] = 0
				continue
			}

			num := element

			if !num.IsNumber() {
				return res.FailAt(1, "List elements must be numbers representing bytes")
			}

			byte64, err := num.AsInt()

			if err != nil {
				return res.Failure(err)
			}

			byteValue := int(byte64)

			if byteValue < 0 || byteValue > 255 {
				return res.FailAt(1, "Byte values must be between 0 and 255")
			}

			bytes[i] = byte(byteValue)
		}

		return res.Success(values.NewBytes(bytes))
	}

	return res.FailAt(1,
		shared.Errors.InvalidArgTypeWithHint("pack", shared.TypeStringOrList, shared.TypeStringOrList))
}
