/*
 *
 * Chippy - internal/builtins/stat.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
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

func statFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("stat", 1, "path"))
	}

	pathStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("stat", shared.TypeString, "path"))
	}

	fileInfo, err := os.Stat(pathStr.Value)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
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
		"size":     values.NewNumber(fileInfo.Size()),
		"modified": values.NewNumber(mtime),
		"accessed": values.NewNumber(atime),
		"changed":  values.NewNumber(ctime),
		"mode":     values.NewNumber(fileModeToChmod(fileInfo.Mode())),
		"userId":   values.NewNumber(uid),
		"groupId":  values.NewNumber(gid),
		"links":    values.NewNumber(nlink),
		"inode":    values.NewNumber(ino),
		"device":   values.NewNumber(dev),
		"type":     values.NewString(fileType),
	}

	result := values.NewMapFromEntries(keys, entries)

	return res.Success(result)
}
