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
func PanicCyclic(val Value, msg string) {
	posStart, posEnd := val.GetPos()

	panic(errors.NewRTError(posStart, posEnd, msg))
}

// EnterWalk guards a walk against cycles and over-deep nesting. Defer the
// returned func to leave: defer EnterWalk(val, seen, depth)().
func EnterWalk(val Value, seen map[Value]bool, depth int) func() {
	if depth > constants.LIMIT_VALUE_NESTING_DEPTH {
		PanicCyclic(val, constants.E_VALUE_TOO_DEEP)
	}

	if seen[val] {
		PanicCyclic(val, constants.E_CYCLIC_VALUE)
	}

	seen[val] = true

	return func() { delete(seen, val) }
}

func walkString(val Value, seen map[Value]bool, depth int) string {
	switch node := val.(type) {
	case nil:
		return "null"
	case *List:
		return node.stringWalk(seen, depth)
	case *Map:
		return node.stringWalk(seen, depth)
	default:
		return val.String()
	}
}

func walkCopy(val Value, seen map[Value]bool, depth int) Value {
	switch node := val.(type) {
	case nil:
		return nil
	case *List:
		return node.copyWalk(seen, depth)
	case *Map:
		return node.copyWalk(seen, depth)
	default:
		return val.Copy()
	}
}

func walkEqual(left, right Value, seen map[Value]bool, depth int) (Value, error) {
	if leftList, ok := left.(*List); ok {
		if _, ok := right.(*List); ok {
			return leftList.eqWalk(right, seen, depth)
		}
	}

	if leftMap, ok := left.(*Map); ok {
		if _, ok := right.(*Map); ok {
			return leftMap.eqWalk(right, seen, depth)
		}
	}

	return left.GetComparisonEe(right)
}
