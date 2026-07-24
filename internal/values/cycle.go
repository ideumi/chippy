/*
 *
 * RR2 - internal/values/cycle.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"chip-go/internal/errors"
)

// Cycle detection for walking values to print, copy or compare. Only lists and
// maps can contain themselves, so each walk tracks the lists and maps it is
// currently inside of and stops if it reaches one again.

// PanicCyclic aborts a walk with a runtime error that is recovered into a
// returned error at the run boundary
func PanicCyclic(v Value, msg string) {
	posStart, posEnd := v.GetPos()

	panic(errors.NewRTError(posStart, posEnd, msg))
}

// EnterWalk guards a walk against cycles and over-deep nesting. Defer the
// returned func to leave: defer EnterWalk(v, seen, depth)().
func EnterWalk(v Value, seen map[Value]bool, depth int) func() {
	if depth > constants.LIMIT_VALUE_NESTING_DEPTH {
		PanicCyclic(v, constants.E_VALUE_TOO_DEEP)
	}

	if seen[v] {
		PanicCyclic(v, constants.E_CYCLIC_VALUE)
	}

	seen[v] = true

	return func() { delete(seen, v) }
}

func walkString(v Value, seen map[Value]bool, depth int) string {
	switch node := v.(type) {
	case nil:
		return "null"
	case *List:
		return node.stringWalk(seen, depth)
	case *Map:
		return node.stringWalk(seen, depth)
	default:
		return v.String()
	}
}

func walkCopy(v Value, seen map[Value]bool, depth int) Value {
	switch node := v.(type) {
	case nil:
		return nil
	case *List:
		return node.copyWalk(seen, depth)
	case *Map:
		return node.copyWalk(seen, depth)
	default:
		return v.Copy()
	}
}

func walkEqual(a, b Value, seen map[Value]bool, depth int) (Value, error) {
	if al, ok := a.(*List); ok {
		if _, ok := b.(*List); ok {
			return al.eqWalk(b, seen, depth)
		}
	}

	if am, ok := a.(*Map); ok {
		if _, ok := b.(*Map); ok {
			return am.eqWalk(b, seen, depth)
		}
	}

	return a.GetComparisonEe(b)
}
