/*
 *
 * Chippy - internal/optional/regex/shared.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"regexp"
	"strconv"
	"unicode/utf8"
)

func byteToRunePosition(text string, byteStart, byteEnd int) (int, int) {
	return utf8.RuneCountInString(text[:byteStart]) + 1, utf8.RuneCountInString(text[:byteEnd])
}

func groupKey(names []string, index int) string {
	if index == 0 {
		return "match"
	}

	if names[index] != "" {
		return names[index]
	}

	return strconv.Itoa(index)
}

func groupPairs(re *regexp.Regexp, match []string) []values.Value {
	names := re.SubexpNames()

	pairs := make([]values.Value, 0, len(match))

	for i, group := range match {
		pairs = append(pairs, values.NewList([]values.Value{
			values.NewString(groupKey(names, i)),
			values.NewString(group),
		}))
	}

	return pairs
}

// Cache admittance gate
func gateExpression(fn string, args []values.Value, expression string, limitIndex int, limitPos string, admittanceIndex int, admittancePos string) (bool, bool, values.RuntimeResult) {
	res := values.NewRuntimeResult()

	limit, err := args[limitIndex].AsInt()

	if err != nil {
		return false, false, res.FailAt(limitIndex+1,
			shared.Errors.InvalidArgTypePositionalWithHint(fn, limitPos, shared.TypeNumber, "limit"))
	}

	admittance, err := args[admittanceIndex].AsInt()

	if err != nil {
		return false, false, res.FailAt(admittanceIndex+1,
			shared.Errors.InvalidArgTypePositionalWithHint(fn, admittancePos, shared.TypeNumber, "admittance"))
	}

	length := int64(len(expression))

	return length <= limit, length <= admittance, res
}
