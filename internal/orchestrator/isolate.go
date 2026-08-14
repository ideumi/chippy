/*
 *
 * Chippy - internal/orchestrator/isolate.go
 *
 */

package orchestrator

import (
	"chip-go/internal/values"
)

// transferState remembers each captured variable already copied, so one captured
// by several functions is copied only once, and each list and map already walked,
// so one that contains itself is walked only once.
type transferState struct {
	cells      map[*values.Value]*values.Value
	containers map[values.Value]bool
}

func (s *transferState) walked(val values.Value) bool {
	if s.containers[val] {
		return true
	}

	s.containers[val] = true

	return false
}

// IsolateForTransfer copies the variables every function captured, so the actor
// receiving them cannot read one while the sender is still writing to it. All
// the items share a single transferState, so a variable captured by more than
// one of them is still a single shared variable on the other side.
func IsolateForTransfer(items ...values.Value) {
	state := &transferState{
		cells:      make(map[*values.Value]*values.Value),
		containers: make(map[values.Value]bool),
	}

	for _, item := range items {
		isolateValue(item, state)
	}
}

func isolateValue(val values.Value, state *transferState) {
	if closure, ok := values.AsBoundaryClosure(val); ok {
		isolateBoundaryClosure(closure, state)

		return
	}

	if list, ok := values.AsList(val); ok {
		if state.walked(val) {
			return
		}

		for _, element := range list.Elements {
			isolateValue(element, state)
		}

		return
	}

	if mapVal, ok := values.AsMap(val); ok {
		if state.walked(val) {
			return
		}

		for _, entryVal := range mapVal.Entries {
			isolateValue(entryVal, state)
		}
	}
}

// Globals are not copied here. They are found by name and attached to the
// receiver's own globals on delivery.
func isolateBoundaryClosure(fn values.BoundaryClosure, state *transferState) {
	old := fn.TransferCells()
	fresh := make([]*values.Value, len(old))

	for i, cell := range old {
		if snap, found := state.cells[cell]; found {
			fresh[i] = snap
			continue
		}

		// The new variable is recorded before its contents are walked,
		// so a capture that leads back to itself finds it already there
		// and stops.
		snap := new(values.Value)
		state.cells[cell] = snap
		fresh[i] = snap

		if captured := *cell; captured.IsSet() {
			copied := captured.Copy()
			*snap = copied
			isolateValue(copied, state)
		}
	}

	fn.SetTransferCells(fresh)
}

// SnapshotUserGlobals makes a deep copy of the globals the user's program
// defined. Builtins and CHIPRT are left out because every actor already has its
// own. A name a function does not define itself is looked up in the globals of
// the actor running it, so a newly spawned actor needs this copy put back there.
func SnapshotUserGlobals(ctx values.Ctx) ([]string, []values.Value) {
	if ctx.Globals == nil {
		return nil, nil
	}

	var names []string
	var snapshot []values.Value

	ctx.Globals.ForEach(func(name string, val values.Value) {
		if name == "CHIPRT" {
			return
		}

		if _, isBuiltin := values.AsBuiltIn(val); isBuiltin {
			return
		}

		names = append(names, name)
		snapshot = append(snapshot, val.Copy())
	})

	return names, snapshot
}

type bindState struct {
	cells      map[*values.Value]bool
	containers map[values.Value]bool
}

func (s *bindState) walked(val values.Value) bool {
	if s.containers[val] {
		return true
	}

	s.containers[val] = true

	return false
}

// Changing which globals a function points at is what makes a function sent to
// another actor find that actor's builtins and CHIPRT rather than the sender's.
func BindValuesToGlobals(items []values.Value, globals values.Ctx) {
	if globals == nil {
		return
	}

	state := &bindState{
		cells:      make(map[*values.Value]bool),
		containers: make(map[values.Value]bool),
	}

	for _, item := range items {
		bindValue(item, globals, state)
	}
}

func bindValue(val values.Value, globals values.Ctx, state *bindState) {
	if closure, ok := values.AsBoundaryClosure(val); ok {
		closure.RebindGlobals(globals)

		for _, cell := range closure.TransferCells() {
			if state.cells[cell] {
				continue
			}

			state.cells[cell] = true
			bindValue(*cell, globals, state)
		}

		return
	}

	if list, ok := values.AsList(val); ok {
		if state.walked(val) {
			return
		}

		for _, element := range list.Elements {
			bindValue(element, globals, state)
		}

		return
	}

	if mapVal, ok := values.AsMap(val); ok {
		if state.walked(val) {
			return
		}

		for _, entryVal := range mapVal.Entries {
			bindValue(entryVal, globals, state)
		}
	}
}
