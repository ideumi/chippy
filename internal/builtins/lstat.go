/*
 *
 * RR2 - internal/builtins/lstat.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
	"syscall"
)

func lstatFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("lstat", 1, "path")))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("lstat", shared.TypeString, "path")))
	}

	fileInfo, err := os.Lstat(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Format: [size, mtime, atime, ctime, mode, uid, gid, nlink, ino, dev, type]

	// Get underlying syscall.Stat_t
	var uid, gid uint32
	var nlink, ino, dev uint64
	atime := fileInfo.ModTime().Unix()
	mtime := fileInfo.ModTime().Unix()
	ctime := fileInfo.ModTime().Unix()

	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		uid = stat.Uid
		gid = stat.Gid
		nlink = uint64(stat.Nlink)
		ino = stat.Ino
		dev = stat.Dev
		atime = stat.Atim.Sec
		mtime = stat.Mtim.Sec
		ctime = stat.Ctim.Sec
	}

	// Determine file type
	var fileType string

	mode := fileInfo.Mode()

	switch {
	case mode.IsRegular():
		fileType = "file"

	case mode.IsDir():
		fileType = "dir"

	case mode&os.ModeSymlink != 0:
		fileType = "symlink"

	case mode&os.ModeDevice != 0:
		if mode&os.ModeCharDevice != 0 {
			fileType = "char"
		} else {
			fileType = "block"
		}

	case mode&os.ModeNamedPipe != 0:
		fileType = "fifo"

	case mode&os.ModeSocket != 0:
		fileType = "socket"

	default:
		fileType = "unknown"
	}

	statElements := []values.Value{
		values.NewNumber(fileInfo.Size()).SetContext(ctx),                  // size
		values.NewNumber(mtime).SetContext(ctx),                            // mtime
		values.NewNumber(atime).SetContext(ctx),                            // atime
		values.NewNumber(ctime).SetContext(ctx),                            // ctime
		values.NewNumber(fileModeToChmod(fileInfo.Mode())).SetContext(ctx), // mode (permissions)
		values.NewNumber(uid).SetContext(ctx),                              // uid
		values.NewNumber(gid).SetContext(ctx),                              // gid
		values.NewNumber(nlink).SetContext(ctx),                            // nlink
		values.NewNumber(ino).SetContext(ctx),                              // inode
		values.NewNumber(dev).SetContext(ctx),                              // device
		values.NewString(fileType).SetContext(ctx),                         // type
	}

	result := values.NewList(statElements)

	return res.Success(result.SetContext(ctx))
}
