/*
 *
 * RR2 - internal/builtins/stat.go
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

func fileModeToChmod(mode os.FileMode) float64 {
	perm := mode.Perm()
	owner := (perm >> 6) & 7
	group := (perm >> 3) & 7
	other := perm & 7

	var special int
	if mode&os.ModeSetuid != 0 {
		special += 4
	}
	if mode&os.ModeSetgid != 0 {
		special += 2
	}
	if mode&os.ModeSticky != 0 {
		special += 1
	}

	return float64(special*1000 + int(owner)*100 + int(group)*10 + int(other))
}

func statFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("stat", 1, "path"),
			ctx,
		))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("stat", shared.TypeString, "path"),
			ctx,
		))
	}

	fileInfo, err := os.Stat(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Format: [size, mtime, atime, ctime, mode, uid, gid, nlink, ino, dev, type]

	// Get underlying syscall.Stat_t
	var uid, gid, nlink, ino, dev float64

	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		uid = float64(stat.Uid)
		gid = float64(stat.Gid)
		nlink = float64(stat.Nlink)
		ino = float64(stat.Ino)
		dev = float64(stat.Dev)
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
		values.NewNumber(float64(fileInfo.Size())).SetContext(ctx),           // size
		values.NewNumber(float64(fileInfo.ModTime().Unix())).SetContext(ctx), // mtime
		values.NewNumber(float64(fileInfo.ModTime().Unix())).SetContext(ctx), // atime (Go doesn't expose separately)
		values.NewNumber(float64(fileInfo.ModTime().Unix())).SetContext(ctx), // ctime (Go doesn't expose separately)
		values.NewNumber(fileModeToChmod(fileInfo.Mode())).SetContext(ctx),   // mode (permissions)
		values.NewNumber(uid).SetContext(ctx),                                // uid
		values.NewNumber(gid).SetContext(ctx),                                // gid
		values.NewNumber(nlink).SetContext(ctx),                              // nlink
		values.NewNumber(ino).SetContext(ctx),                                // inode
		values.NewNumber(dev).SetContext(ctx),                                // device
		values.NewString(fileType).SetContext(ctx),                           // type
	}

	result := values.NewList(statElements)
	return res.Success(result.SetContext(ctx))
}
