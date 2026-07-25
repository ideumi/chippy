/*
 *
 * RR2 - internal/builtins/rand.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"crypto/rand"
)

func randFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("rand", 1, "count"))
	}

	bytesNum, ok := args[0].(*values.Number)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("rand", shared.TypeNumber, "count"))
	}

	count64, err := bytesNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	numBytes := int(count64)

	if numBytes <= 0 {
		return res.FailAt(1, shared.Errors.InvalidValue("Byte count must be positive"))
	}

	buf := make([]byte, numBytes)
	_, err = rand.Read(buf)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewBytes(buf).SetContext(ctx))
}
