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

func fileModeToChmod(mode os.FileMode) int {
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

	return special*1000 + int(owner)*100 + int(group)*10 + int(other)
}

func statFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("stat", 1, "path")))
	}

	pathStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("stat", shared.TypeString, "path")))
	}

	fileInfo, err := os.Stat(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

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

	keys := []string{
		"size",
		"modified",
		"accessed",
		"changed",
		"mode",
		"userId",
		"groupId",
		"links",
		"inode",
		"device",
		"type",
	}

	entries := map[string]values.Value{
		"size":     values.NewNumber(fileInfo.Size()).SetContext(ctx),
		"modified": values.NewNumber(mtime).SetContext(ctx),
		"accessed": values.NewNumber(atime).SetContext(ctx),
		"changed":  values.NewNumber(ctime).SetContext(ctx),
		"mode":     values.NewNumber(fileModeToChmod(fileInfo.Mode())).SetContext(ctx),
		"userId":   values.NewNumber(uid).SetContext(ctx),
		"groupId":  values.NewNumber(gid).SetContext(ctx),
		"links":    values.NewNumber(nlink).SetContext(ctx),
		"inode":    values.NewNumber(ino).SetContext(ctx),
		"device":   values.NewNumber(dev).SetContext(ctx),
		"type":     values.NewString(fileType).SetContext(ctx),
	}

	result := values.NewMapFromEntries(keys, entries)

	return res.Success(result.SetContext(ctx))
}
