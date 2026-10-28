/*
 *
 * Chippy - internal/optional/regex/findallindexes.go
 *
 */

package regex

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/values"
	"unicode/utf8"
)

func findAllIndexesFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 4 {
		return res.Fail(
			shared.Errors.InvalidArgCountWithHint(optional.Prefixed(OptionalName, "findallindexes"), 4, "expression, text, limit, admittance"))
	}

	expression, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "findallindexes"), shared.PositionFirst, shared.TypeString, "expression"))
	}

	text, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint(optional.Prefixed(OptionalName, "findallindexes"), shared.PositionSecond, shared.TypeString, "text"))
	}

	ok, admitted, gate := gateExpression(optional.Prefixed(OptionalName, "findallindexes"), args, expression.Value, 2, shared.PositionThird, 3, shared.PositionFourth)

	if gate.Error != nil {
		return gate
	}

	if !ok {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	re, err := getCompiled(expression.Value, admitted)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	spans := re.FindAllStringIndex(text.Value, -1)

	elements := make([]values.Value, 0, len(spans))

	// Rune positions are accumulated in a single pass. Each span extends the
	// running count from where the previous span ended, so every byte of the
	// text is scanned at most once, no matter how many matches there are.
	runeCount := 0
	byteBase := 0

	for _, span := range spans {
		byteStart, byteEnd := span[0], span[1]

		runeCount += utf8.RuneCountInString(text.Value[byteBase:byteStart])
		start := runeCount + 1

		runeCount += utf8.RuneCountInString(text.Value[byteStart:byteEnd])
		end := runeCount

		byteBase = byteEnd

		elements = append(elements, values.NewMapFromEntries(
			[]string{"start", "end"},
			map[string]values.Value{
				"start": values.NewNumber(start),
				"end":   values.NewNumber(end),
			}))
	}

	return res.Success(values.NewList(elements))
}
