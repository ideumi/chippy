/*
 *
 * RR2 - internal/builtins/sort.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"sort"
	"strings"
)

func sortFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sort", 1, "list"))
	}

	listArg, ok := args[0].(*values.List)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("sort", shared.TypeList, "list"))
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
func compareValues(left, right values.Value, seen map[values.Value]bool, depth int) int {
	leftType := getTypePrecedence(left)
	rightType := getTypePrecedence(right)

	// Different types: compare by precedence
	if leftType != rightType {
		return leftType - rightType
	}

	// Same type: natural comparison
	switch leftVal := left.(type) {
	case *values.Number:
		rightVal := right.(*values.Number)

		if leftVal.IsInt() && rightVal.IsInt() {
			leftInt, _ := leftVal.AsInt()
			rightInt, _ := rightVal.AsInt()

			if leftInt < rightInt {
				return -1
			}

			if leftInt > rightInt {
				return 1
			}

			return 0
		}

		if leftVal.AsFloat() < rightVal.AsFloat() {
			return -1
		}

		if leftVal.AsFloat() > rightVal.AsFloat() {
			return 1
		}

		return 0

	case *values.String:
		rightVal := right.(*values.String)

		return strings.Compare(leftVal.Value, rightVal.Value)

	case *values.List:
		rightVal := right.(*values.List)

		return compareLists(leftVal, rightVal, seen, depth)

	case *values.Bytes:
		rightVal := right.(*values.Bytes)

		return compareBytes(leftVal.Data, rightVal.Data)

	default:
		// Use string representation for everything else
		return strings.Compare(left.String(), right.String())
	}
}

// getTypePrecedence returns precedence value for sorting
func getTypePrecedence(val values.Value) int {
	switch val.(type) {
	case *values.Number:
		return 0
	case *values.String:
		return 1
	case *values.List:
		return 2
	case *values.Bytes:
		return 3
	case *values.BuiltInFunction, values.Callable:
		return 4
	case *values.Map:
		return 5
	default:
		return 6
	}
}

// compareLists compares two lists element by element.
func compareLists(left, right *values.List, seen map[values.Value]bool, depth int) int {
	// Catch cyclic lists
	defer values.EnterWalk(left, seen, depth)()

	minLen := len(left.Elements)

	if len(right.Elements) < minLen {
		minLen = len(right.Elements)
	}

	for i := 0; i < minLen; i++ {
		cmp := compareValues(left.Elements[i], right.Elements[i], seen, depth+1)

		if cmp != 0 {
			return cmp
		}
	}

	// If all compared elements are equal, shorter list comes first
	return len(left.Elements) - len(right.Elements)
}

// compareBytes compares two byte slices
func compareBytes(left, right []byte) int {
	minLen := len(left)

	if len(right) < minLen {
		minLen = len(right)
	}

	for i := 0; i < minLen; i++ {
		if left[i] < right[i] {
			return -1
		} else if left[i] > right[i] {
			return 1
		}
	}

	return len(left) - len(right)
}
