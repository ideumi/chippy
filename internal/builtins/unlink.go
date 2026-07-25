/*
 *
 * RR2 - internal/builtins/unlink.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
)

func unlinkFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("unlink", 1, "path"))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("unlink", shared.TypeString, "path"))
	}

	err := os.Remove(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
