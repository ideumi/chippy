/*
 *
 * Chippy - internal/orchestrator/receive.go
 *
 */

package orchestrator

import (
	"chip-go/internal/values"
)

func (o *Orchestrator) tryDrainAndRun(inst *Instance) []values.Value {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst.Inbox.mu.Lock()
	defer inst.Inbox.mu.Unlock()

	if len(inst.Inbox.items) == 0 {
		return nil
	}

	items := inst.Inbox.items
	inst.Inbox.items = make([]values.Value, 0)

	inst.State = StateRunning

	clearBlocking(inst)

	return items
}

// ReceiveBlocking delivers queued messages to inst, blocking until items arrive
// or the deadlock detector cancels the actor. Returns (nil, true) on cancel.
// Owns the Begin/EndBlocking lifecycle so transitions stay atomic with inbox
// drains.
func (o *Orchestrator) ReceiveBlocking(inst *Instance, globals values.Ctx) ([]values.Value, bool) {
	if items := o.tryDrainAndRun(inst); items != nil {
		BindValuesToGlobals(items, globals)

		return items, false
	}

	cancelCh := o.BeginBlocking(inst, inst.Inbox.HasItems)

	for {
		select {
		case <-inst.Inbox.signal:
			if items := o.tryDrainAndRun(inst); items != nil {
				BindValuesToGlobals(items, globals)
				return items, false
			}

		case <-cancelCh:
			if items := o.tryDrainAndRun(inst); items != nil {
				BindValuesToGlobals(items, globals)
				return items, false
			}

			o.EndBlocking(inst)

			return nil, true
		}
	}
}
