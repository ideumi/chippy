/*
 *
 * RR2 - internal/builtins/rand.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"crypto/rand"
)

func randFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("rand", 1, "count"),
			ctx,
		))
	}

	bytesNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("rand", shared.TypeNumber, "count"),
			ctx,
		))
	}

	numBytes := int(bytesNum.Value)

	if numBytes <= 0 {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Byte count must be positive"),
			ctx,
		))
	}

	buf := make([]byte, numBytes)
	_, err := rand.Read(buf)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	var byteValues []values.Value
	for _, b := range buf {
		byteValues = append(byteValues, values.NewNumber(float64(b)).SetContext(ctx))
	}

	return res.Success(values.NewList(byteValues).SetContext(ctx))
}
