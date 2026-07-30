/*
 *
 * RR2 - internal/builtins/rename.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
)

func renameFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("rename", 2, "old path, new path"))
	}

	oldPathStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("rename", shared.PositionFirst, shared.TypeString, "old path"))
	}

	newPathStr, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("rename", shared.PositionSecond, shared.TypeString, "new path"))
	}

	// Rename/move file or directory
	err := os.Rename(oldPathStr.Value, newPathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(constants.STR_OK))
}
