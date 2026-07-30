/*
 *
 * Chippy - internal/builtins/append.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func appendFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("append", 2, "list, value"))
	}

	value := args[0]

	if bytesVal, ok := values.AsBytes(value); ok {
		valueNum := args[1]

		if !valueNum.IsNumber() {
			return res.FailAt(2,
				shared.Errors.InvalidArgTypePositionalWithHint("append", shared.PositionSecond, shared.TypeNumber, "value"))
		}

		byte64, err := valueNum.AsInt()

		if err != nil {
			return res.Failure(err)
		}

		byteValue := int(byte64)

		if byteValue < 0 || byteValue > 255 {
			return res.FailAt(2, "Byte values must be between 0 and 255")
		}

		bytesVal.AppendByte(byteValue)

		return res.Success(value)
	}

	if listVal, ok := values.AsList(value); ok {
		listVal.Elements = append(listVal.Elements, args[1])

		return res.Success(value)
	}

	return res.FailAt(1,
		shared.Errors.InvalidArgTypePositionalWithHint("append", shared.PositionFirst, shared.TypeListOrBytes, shared.TypeListOrBytes))
}
