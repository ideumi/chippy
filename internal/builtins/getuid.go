/*
 *
 * RR2 - internal/builtins/getuid.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
)

func getuidFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("getuid", 0),
			ctx,
		))
	}

	uid := os.Getuid()

	return res.Success(values.NewNumber(float64(uid)).SetContext(ctx))
}
