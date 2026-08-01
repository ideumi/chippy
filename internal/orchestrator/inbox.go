/*
 *
 * Chippy - internal/orchestrator/inbox.go
 *
 */

package orchestrator

import (
	"chip-go/internal/values"
	"sync"
)

// When both locks are needed, Orchestrator.mu is taken first and Inbox.mu second.
// Anything that takes them the other way round can deadlock against this.
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

// Send takes val over. The caller must not use it again afterwards, because any
// function inside it has the variables it captured replaced with copies and no
// longer points at the ones the caller is still using.
func (in *Inbox) Send(val values.Value) {
	IsolateForTransfer(val)

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
