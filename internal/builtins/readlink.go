/*
 *
 * RR2 - internal/builtins/readlink.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
)

func readlinkFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("readlink", 1, "path"))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("readlink", shared.TypeString, "path"))
	}

	target, err := os.Readlink(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(target).SetContext(ctx))
}
