/*
 *
 * RR2 - internal/builtins/symlink.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
)

func symlinkFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("symlink", 2, "target, link path"))
	}

	targetStr, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("symlink", shared.PositionFirst, shared.TypeString, "target"))
	}

	linkPathStr, ok := args[1].(*values.String)

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("symlink", shared.PositionSecond, shared.TypeString, "link path"))
	}

	err := os.Symlink(targetStr.Value, linkPathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
