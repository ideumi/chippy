/*
 *
 * RR2 - internal/builtins/sort.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"sort"
	"strings"
)

func sortFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("sort", 1, "list")))
	}

	listArg, ok := args[0].(*values.List)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("sort", shared.TypeList, "list")))
	}

	// Reuse seen across all compares
	// EnterWalk's defer empties it between them.
	seen := map[values.Value]bool{}

	sort.SliceStable(listArg.Elements, func(i, j int) bool {
		return compareValues(listArg.Elements[i], listArg.Elements[j], seen, 0) < 0
	})

	return res.Success(listArg.SetContext(ctx))
}

// compareValues implements type-aware comparison with precedence:
// Number < String < List < Bytes < Function < Map
func compareValues(a, b values.Value, seen map[values.Value]bool, depth int) int {
	aType := getTypePrecedence(a)
	bType := getTypePrecedence(b)

	// Different types: compare by precedence
	if aType != bType {
		return aType - bType
	}

	// Same type: natural comparison
	switch aVal := a.(type) {
	case *values.Number:
		bVal := b.(*values.Number)

		if aVal.IsInt() && bVal.IsInt() {
			aInt, _ := aVal.AsInt()
			bInt, _ := bVal.AsInt()

			if aInt < bInt {
				return -1
			}

			if aInt > bInt {
				return 1
			}

			return 0
		}

		if aVal.AsFloat() < bVal.AsFloat() {
			return -1
		}

		if aVal.AsFloat() > bVal.AsFloat() {
			return 1
		}

		return 0

	case *values.String:
		bVal := b.(*values.String)

		return strings.Compare(aVal.Value, bVal.Value)

	case *values.List:
		bVal := b.(*values.List)

		return compareLists(aVal, bVal, seen, depth)

	case *values.Bytes:
		bVal := b.(*values.Bytes)

		return compareBytes(aVal.Data, bVal.Data)

	default:
		// Use string representation for everything else
		return strings.Compare(a.String(), b.String())
	}
}

// getTypePrecedence returns precedence value for sorting
func getTypePrecedence(v values.Value) int {
	switch v.(type) {
	case *values.Number:
		return 0
	case *values.String:
		return 1
	case *values.List:
		return 2
	case *values.Bytes:
		return 3
	case *values.BuiltInFunction, *values.Function:
		return 4
	case *values.Map:
		return 5
	default:
		return 6
	}
}

// compareLists compares two lists element by element.
func compareLists(a, b *values.List, seen map[values.Value]bool, depth int) int {
	// Catch cyclic lists
	defer values.EnterWalk(a, seen, depth)()

	minLen := len(a.Elements)

	if len(b.Elements) < minLen {
		minLen = len(b.Elements)
	}

	for i := 0; i < minLen; i++ {
		cmp := compareValues(a.Elements[i], b.Elements[i], seen, depth+1)

		if cmp != 0 {
			return cmp
		}
	}

	// If all compared elements are equal, shorter list comes first
	return len(a.Elements) - len(b.Elements)
}

// compareBytes compares two byte slices
func compareBytes(a, b []byte) int {
	minLen := len(a)

	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}

	return len(a) - len(b)
}
