/*
 *
 * RR2 - internal/builtins/shared/dirhandles.go
 *
 */

package shared

import (
	"os"
	"sync"
)

type DirectoryHandle struct {
	ID      int
	DirFile *os.File
	Path    string
}

var (
	dirHandles      = make(map[int]*DirectoryHandle)
	dirHandleMutex  sync.RWMutex
	nextDirHandleID = 1
)

// AddDirHandle adds a new directory handle and returns its ID
func AddDirHandle(dirFile *os.File, path string) int {

	// Use the shared handle recycling system instead of separate counter
	id := GetNextFileHandle()

	dirHandleMutex.Lock()
	defer dirHandleMutex.Unlock()

	dirHandles[id] = &DirectoryHandle{
		ID:      id,
		DirFile: dirFile,
		Path:    path,
	}

	return id
}

func GetDirHandle(id int) (*DirectoryHandle, bool) {
	dirHandleMutex.RLock()

	defer dirHandleMutex.RUnlock()

	handle, exists := dirHandles[id]

	return handle, exists
}

func RemoveDirHandle(id int) {
	dirHandleMutex.Lock()
	handle, exists := dirHandles[id]

	if exists {
		handle.DirFile.Close()
		delete(dirHandles, id)
	}

	dirHandleMutex.Unlock()

	if exists {
		// Recycle handle ID
		RecycleFileHandle(id)
	}
}
