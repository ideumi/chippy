/*
 *
 * RR2 - internal/handles/processhandles.go
 *
 */

package handles

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

type ProcessHandles struct {
	handles map[int]*ProcessHandle
	mu      sync.RWMutex
}

func NewProcessHandles() *ProcessHandles {
	return &ProcessHandles{
		handles: make(map[int]*ProcessHandle),
	}
}

func (ph *ProcessHandles) Get(id int) (*ProcessHandle, bool) {
	ph.mu.RLock()
	defer ph.mu.RUnlock()
	proc, ok := ph.handles[id]

	return proc, ok
}

func (ph *ProcessHandles) Store(id int, proc *ProcessHandle) {
	ph.mu.Lock()
	defer ph.mu.Unlock()
	ph.handles[id] = proc
}

func (ph *ProcessHandles) Remove(id int) {
	ph.mu.Lock()
	defer ph.mu.Unlock()
	delete(ph.handles, id)
}

func (ph *ProcessHandles) Extract(id int) (*ProcessHandle, bool) {
	ph.mu.Lock()
	defer ph.mu.Unlock()
	proc, ok := ph.handles[id]

	if ok {
		delete(ph.handles, id)
	}

	return proc, ok
}

func (ph *ProcessHandles) CloseAll() {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	for id, proc := range ph.handles {
		closeProcess(proc)
		delete(ph.handles, id)
	}
}

func closeProcess(process *ProcessHandle) {
	if process.Cmd != nil && process.Cmd.Process != nil {
		process.Cmd.Process.Kill()
	}

	if process.Stdin != nil {
		process.Stdin.Close()
	}

	if process.Stdout != nil {
		process.Stdout.Close()
	}

	if process.Stderr != nil {
		process.Stderr.Close()
	}

	if process.Cmd != nil {
		process.Cmd.Wait()
	}
}
