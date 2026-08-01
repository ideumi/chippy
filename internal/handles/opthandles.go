/*
 *
 * Chippy - internal/handles/opthandles.go
 *
 */

package handles

import "sync"

type OptHandles struct {
	handles map[int]interface{}
	mu      sync.RWMutex
}

func NewOptHandles() *OptHandles {
	return &OptHandles{
		handles: make(map[int]interface{}),
	}
}

func (oh *OptHandles) Get(id int) (interface{}, bool) {
	oh.mu.RLock()
	defer oh.mu.RUnlock()
	opt, ok := oh.handles[id]

	return opt, ok
}

func (oh *OptHandles) Store(id int, opt interface{}) {
	oh.mu.Lock()
	defer oh.mu.Unlock()
	oh.handles[id] = opt
}

func (oh *OptHandles) Remove(id int) {
	oh.mu.Lock()
	defer oh.mu.Unlock()
	delete(oh.handles, id)
}

func (oh *OptHandles) Extract(id int) (interface{}, bool) {
	oh.mu.Lock()
	defer oh.mu.Unlock()
	opt, ok := oh.handles[id]

	if ok {
		delete(oh.handles, id)
	}

	return opt, ok
}

func (oh *OptHandles) CloseAll() {
	oh.mu.Lock()
	defer oh.mu.Unlock()

	for id, opt := range oh.handles {
		if closer, ok := opt.(interface{ Close() error }); ok {
			closer.Close()
		}

		delete(oh.handles, id)
	}
}
