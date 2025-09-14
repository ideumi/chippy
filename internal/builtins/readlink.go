/*
 *
 * RR2 - internal/builtins/readlink.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
)

func readlinkFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("readlink", 1, "path"),
			ctx,
		))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("readlink", shared.TypeString, "path"),
			ctx,
		))
	}

	target, err := os.Readlink(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(target).SetContext(ctx))
}
