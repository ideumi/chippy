/*
 *
 * RR2 - internal/handles/filehandles.go
 *
 */

package handles

import (
	"os"
	"sync"
)

type FileHandles struct {
	handles map[int]*os.File
	mu      sync.RWMutex
}

func NewFileHandles() *FileHandles {
	fh := &FileHandles{
		handles: make(map[int]*os.File),
	}

	fh.handles[0] = os.Stdin
	fh.handles[1] = os.Stdout
	fh.handles[2] = os.Stderr

	return fh
}

func (fh *FileHandles) Get(id int) (*os.File, bool) {
	fh.mu.RLock()
	defer fh.mu.RUnlock()
	file, ok := fh.handles[id]

	return file, ok
}

func (fh *FileHandles) Store(id int, file *os.File) {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	fh.handles[id] = file
}

func (fh *FileHandles) Remove(id int) {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	delete(fh.handles, id)
}

func (fh *FileHandles) Extract(id int) (*os.File, bool) {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	file, ok := fh.handles[id]

	if ok {
		delete(fh.handles, id)
	}

	return file, ok
}

func (fh *FileHandles) CloseAll() {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	for id, file := range fh.handles {
		if id > StdioMaxID {
			file.Close()
			delete(fh.handles, id)
		}
	}
}
