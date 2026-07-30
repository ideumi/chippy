/*
 *
 * RR2 - internal/builtins/dopen.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/handles"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os"
)

func dopenFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("dopen", 1, "path"))
	}

	pathStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("dopen", shared.TypeString, "path"))
	}

	dirFile, err := os.Open(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	// Verify its actually a directory
	stat, err := dirFile.Stat()

	if err != nil || !stat.IsDir() {
		dirFile.Close()

		return res.Success(values.NewString(constants.STR_ERR))
	}

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	handle := registry.Alloc.Alloc()
	registry.Dirs.Store(handle, &handles.DirectoryHandle{
		DirFile: dirFile,
		Path:    pathStr.Value,
	})

	return res.Success(values.NewNumber(handle))
}
