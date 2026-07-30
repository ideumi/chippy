/*
 *
 * Chippy - internal/builtins/sort.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"sort"
	"strings"
)

func sortFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sort", 1, "list"))
	}

	listArg, ok := values.AsList(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("sort", shared.TypeList, "list"))
	}

	// Reuse seen across all compares
	// EnterWalk's defer empties it between them.
	seen := map[values.Value]bool{}

	sort.SliceStable(listArg.Elements, func(i, j int) bool {
		return compareValues(listArg.Elements[i], listArg.Elements[j], seen, 0) < 0
	})

	return res.Success(args[0])
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
	if left.IsNumber() {
		if left.IsInt() && right.IsInt() {
			leftInt, _ := left.AsInt()
			rightInt, _ := right.AsInt()

			if leftInt < rightInt {
				return -1
			}

			if leftInt > rightInt {
				return 1
			}

			return 0
		}

		if left.AsFloat() < right.AsFloat() {
			return -1
		}

		if left.AsFloat() > right.AsFloat() {
			return 1
		}

		return 0
	}

	if leftStr, ok := values.AsString(left); ok {
		rightStr, _ := values.AsString(right)

		return strings.Compare(leftStr.Value, rightStr.Value)
	}

	if _, ok := values.AsList(left); ok {
		return compareLists(left, right, seen, depth)
	}

	if leftBytes, ok := values.AsBytes(left); ok {
		rightBytes, _ := values.AsBytes(right)

		return compareBytes(leftBytes.Data, rightBytes.Data)
	}

	// Use string representation for everything else
	return strings.Compare(left.String(), right.String())
}

// getTypePrecedence returns precedence value for sorting
func getTypePrecedence(val values.Value) int {
	if val.IsNumber() {
		return 0
	}

	if _, ok := values.AsString(val); ok {
		return 1
	}

	if _, ok := values.AsList(val); ok {
		return 2
	}

	if _, ok := values.AsBytes(val); ok {
		return 3
	}

	if _, ok := values.AsBuiltIn(val); ok {
		return 4
	}

	if _, ok := values.AsCallable(val); ok {
		return 4
	}

	if _, ok := values.AsMap(val); ok {
		return 5
	}

	return 6
}

// compareLists compares two lists element by element.
func compareLists(left, right values.Value, seen map[values.Value]bool, depth int) int {
	// Catch cyclic lists
	defer values.EnterWalk(left, seen, depth)()

	leftList, _ := values.AsList(left)
	rightList, _ := values.AsList(right)

	minLen := len(leftList.Elements)

	if len(rightList.Elements) < minLen {
		minLen = len(rightList.Elements)
	}

	for i := 0; i < minLen; i++ {
		cmp := compareValues(leftList.Elements[i], rightList.Elements[i], seen, depth+1)

		if cmp != 0 {
			return cmp
		}
	}

	// If all compared elements are equal, shorter list comes first
	return len(leftList.Elements) - len(rightList.Elements)
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
