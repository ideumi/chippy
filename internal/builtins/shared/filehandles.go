/*
 *
 * RR2 - internal/builtins/shared/filehandles.go
 *
 */

package shared

import (
	"os"
	"sync"
)

var (
	fileHandles  = make(map[int]*os.File)
	nextHandle   = 3
	freedHandles = make([]int, 0) // Pool of freed handle IDs for reuse
	handleMutex  sync.RWMutex
)

func init() {
	fileHandles[0] = os.Stdin
	fileHandles[1] = os.Stdout
	fileHandles[2] = os.Stderr
}

// GetNextFileHandle returns either a recycled handle ID or allocates a new one
func GetNextFileHandle() int {
	handleMutex.Lock()
	defer handleMutex.Unlock()

	// Reuse freed handle if available
	if len(freedHandles) > 0 {
		handle := freedHandles[len(freedHandles)-1]
		freedHandles = freedHandles[:len(freedHandles)-1]
		return handle
	}

	// Otherwise allocate new handle
	handle := nextHandle
	nextHandle++
	return handle
}

// RecycleFileHandle adds a handle ID back to the freed pool
func RecycleFileHandle(handle int) {
	handleMutex.Lock()

	defer handleMutex.Unlock()

	// Only recycle handles > 2 (don't recycle stdin/stdout/stderr)
	if handle > 2 {
		freedHandles = append(freedHandles, handle)
	}
}

// GetFileHandle gets a file handle
func GetFileHandle(handle int) (*os.File, bool) {
	handleMutex.RLock()
	defer handleMutex.RUnlock()
	file, exists := fileHandles[handle]
	return file, exists
}

// StoreFileHandle stores a file handle
func StoreFileHandle(handle int, file *os.File) {
	handleMutex.Lock()
	defer handleMutex.Unlock()
	fileHandles[handle] = file
}

// RemoveFileHandle removes a file handle
func RemoveFileHandle(handle int) {
	handleMutex.Lock()
	defer handleMutex.Unlock()
	delete(fileHandles, handle)
}
