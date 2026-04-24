/*
 *
 * RR2 - internal/handles/alloc.go
 *
 */

package handles

import "sync"

const StdioMaxID = 2 // 0, 1, 2 are standard handles

type HandleAllocator struct {
	nextHandle   int
	freedHandles []int
	mu           sync.Mutex
}

func NewHandleAllocator(startID int) *HandleAllocator {
	return &HandleAllocator{
		nextHandle:   startID,
		freedHandles: make([]int, 0),
	}
}

func (a *HandleAllocator) Alloc() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.freedHandles) > 0 {
		id := a.freedHandles[len(a.freedHandles)-1]
		a.freedHandles = a.freedHandles[:len(a.freedHandles)-1]

		return id
	}

	id := a.nextHandle
	a.nextHandle++

	return id
}

func (a *HandleAllocator) Free(id int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if id > StdioMaxID {
		a.freedHandles = append(a.freedHandles, id)
	}
}
