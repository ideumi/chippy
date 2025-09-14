/*
 *
 * RR2 - internal/builtins/getppid.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
)

func getppidFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("getppid", 0),
			ctx,
		))
	}

	ppid := os.Getppid()

	return res.Success(values.NewNumber(float64(ppid)).SetContext(ctx))
}
