/*
 *
 * RR2 - internal/orchestrator/isolate.go
 *
 */

package orchestrator

import (
	"chip-go/internal/values"
)

// transferState remembers each captured variable already copied, so one captured
// by several functions is copied only once. Cycle handling lives at the point it
// matters, in isolateBoundaryClosure.
type transferState struct {
	cells map[*values.Value]*values.Value
}

// IsolateForTransfer copies the variables every function captured, so the actor
// receiving them cannot read one while the sender is still writing to it. All
// the items share a single transferState, so a variable captured by more than
// one of them is still a single shared variable on the other side.
func IsolateForTransfer(items ...values.Value) {
	state := &transferState{
		cells: make(map[*values.Value]*values.Value),
	}

	for _, item := range items {
		isolateValue(item, state)
	}
}

func isolateValue(val values.Value, state *transferState) {
	switch typed := val.(type) {
	case values.BoundaryClosure:
		isolateBoundaryClosure(typed, state)
	case *values.List:
		for _, element := range typed.Elements {
			isolateValue(element, state)
		}
	case *values.Map:
		for _, entryVal := range typed.Entries {
			isolateValue(entryVal, state)
		}
	}
}

// Every captured variable is replaced by a new one holding a copy, so the receiver
// shares nothing with the variables the sender is still using. Globals are not
// copied here. They are found by name and attached to the receiver's own globals
// on delivery.
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

		if captured := *cell; captured != nil {
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

		if _, isBuiltin := val.(*values.BuiltInFunction); isBuiltin {
			return
		}

		names = append(names, name)
		snapshot = append(snapshot, val.Copy())
	})

	return names, snapshot
}

// Functions are skipped here. The variables they captured are looked after by
// IsolateForTransfer and BindValuesToGlobals instead.
func DeepRebindContext(val values.Value, ctx values.Ctx) {
	if val == nil {
		return
	}

	switch typed := val.(type) {
	case values.Callable, *values.BuiltInFunction:

	case *values.List:
		val.SetContext(ctx)

		for _, element := range typed.Elements {
			DeepRebindContext(element, ctx)
		}

	case *values.Map:
		val.SetContext(ctx)

		for _, entryVal := range typed.Entries {
			DeepRebindContext(entryVal, ctx)
		}

	default:
		val.SetContext(ctx)
	}
}

// bindState remembers the captured variables already dealt with, so one shared
// by several functions, or one that leads back to itself, is handled only once.
type bindState struct {
	cells map[*values.Value]bool
}

// Changing which globals a function points at is what makes a function sent to
// another actor find that actor's builtins and CHIPRT rather than the sender's.
func BindValuesToGlobals(items []values.Value, globals values.Ctx) {
	if globals == nil {
		return
	}

	state := &bindState{
		cells: make(map[*values.Value]bool),
	}

	for _, item := range items {
		bindValue(item, globals, state)
	}
}

func bindValue(val values.Value, globals values.Ctx, state *bindState) {
	switch typed := val.(type) {
	case values.BoundaryClosure:
		typed.RebindGlobals(globals)

		for _, cell := range typed.TransferCells() {
			if state.cells[cell] {
				continue
			}

			state.cells[cell] = true
			bindValue(*cell, globals, state)
		}

	case *values.List:
		for _, element := range typed.Elements {
			bindValue(element, globals, state)
		}

	case *values.Map:
		for _, entryVal := range typed.Entries {
			bindValue(entryVal, globals, state)
		}
	}
}
