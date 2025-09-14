/*
 *
 * RR2 - internal/builtins/dopen.go
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

func dopenFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("dopen", 1, "path"),
			ctx,
		))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("dopen", shared.TypeString, "path"),
			ctx,
		))
	}

	dirFile, err := os.Open(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Verify its actually a directory
	stat, err := dirFile.Stat()

	if err != nil || !stat.IsDir() {
		dirFile.Close()

		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Add to directory handle registry
	handleID := shared.AddDirHandle(dirFile, pathStr.Value)

	return res.Success(values.NewNumber(float64(handleID)).SetContext(ctx))
}
