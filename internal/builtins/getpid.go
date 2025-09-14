/*
 *
 * RR2 - internal/builtins/getpid.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
)

func getpidFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("getpid", 0),
			ctx,
		))
	}

	pid := os.Getpid()

	return res.Success(values.NewNumber(float64(pid)).SetContext(ctx))
}
