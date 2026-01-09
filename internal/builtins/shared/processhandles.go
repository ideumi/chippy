/*
 *
 * RR2 - internal/builtins/shared/processhandles.go
 *
 */

package shared

import (
	"io"
	"os/exec"
	"sync"
)

type ProcessHandle struct {
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

var (
	processHandles     = make(map[int]*ProcessHandle)
	processHandleMutex sync.RWMutex
)

// GetProcessHandle gets a process handle
func GetProcessHandle(handle int) (*ProcessHandle, bool) {
	processHandleMutex.RLock()
	defer processHandleMutex.RUnlock()
	proc, exists := processHandles[handle]

	return proc, exists
}

// StoreProcessHandle stores a process handle
func StoreProcessHandle(handle int, proc *ProcessHandle) {
	processHandleMutex.Lock()
	defer processHandleMutex.Unlock()
	processHandles[handle] = proc
}

// RemoveProcessHandle removes a process handle
func RemoveProcessHandle(handle int) {
	processHandleMutex.Lock()
	defer processHandleMutex.Unlock()
	delete(processHandles, handle)
}
