/*
 *
 * RR2 - internal/handles/dirhandles.go
 *
 */

package handles

import (
	"os"
	"sync"
)

type DirectoryHandle struct {
	DirFile *os.File
	Path    string
}

type DirHandles struct {
	handles map[int]*DirectoryHandle
	mu      sync.RWMutex
}

func NewDirHandles() *DirHandles {
	return &DirHandles{
		handles: make(map[int]*DirectoryHandle),
	}
}

func (dh *DirHandles) Get(id int) (*DirectoryHandle, bool) {
	dh.mu.RLock()
	defer dh.mu.RUnlock()
	handle, ok := dh.handles[id]

	return handle, ok
}

func (dh *DirHandles) Store(id int, handle *DirectoryHandle) {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	dh.handles[id] = handle
}

func (dh *DirHandles) Remove(id int) {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	delete(dh.handles, id)
}

func (dh *DirHandles) Extract(id int) (*DirectoryHandle, bool) {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	handle, ok := dh.handles[id]

	if ok {
		delete(dh.handles, id)
	}

	return handle, ok
}

func (dh *DirHandles) CloseAll() {
	dh.mu.Lock()
	defer dh.mu.Unlock()

	for id, handle := range dh.handles {
		handle.DirFile.Close()
		delete(dh.handles, id)
	}
}
