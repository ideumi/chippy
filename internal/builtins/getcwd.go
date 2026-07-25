/*
 *
 * RR2 - internal/builtins/getcwd.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
)

func getcwdFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("getcwd", 0))
	}

	cwd, err := os.Getwd()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(cwd).SetContext(ctx))
}
