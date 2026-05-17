/*
 *
 * RR2 - internal/orchestrator/inbox.go
 *
 */

package orchestrator

import (
	"chip-go/internal/values"
	"sync"
)

// Lock order: Orchestrator.mu before Inbox.mu.
type Inbox struct {
	mu     sync.Mutex
	items  []values.Value
	signal chan struct{}
}

func NewInbox() *Inbox {
	return &Inbox{
		items:  make([]values.Value, 0),
		signal: make(chan struct{}, 1),
	}
}

func (in *Inbox) HasItems() bool {
	in.mu.Lock()
	defer in.mu.Unlock()

	return len(in.items) > 0
}

// Send takes ownership of val: callers must not touch val afterwards because
// isolation rebinds any closures inside it.
func (in *Inbox) Send(val values.Value) {
	IsolateForTransfer(val, make(map[values.Ctx]values.Ctx))

	in.mu.Lock()
	in.items = append(in.items, val)
	in.mu.Unlock()

	select {
	case in.signal <- struct{}{}:
	default:
	}
}

func (in *Inbox) drain() []values.Value {
	in.mu.Lock()
	defer in.mu.Unlock()

	if len(in.items) == 0 {
		return nil
	}

	items := in.items
	in.items = make([]values.Value, 0)

	return items
}

func (in *Inbox) ReceiveNonBlocking(globals values.Ctx) []values.Value {
	items := in.drain()

	if items == nil {
		return []values.Value{}
	}

	BindValuesToGlobals(items, globals)

	return items
}
