/*
 *
 * Chippy - internal/builtins/timezone.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"time"
)

func timezoneFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("timezone", 1, "timestamp"))
	}

	tsNum := args[0]

	if !tsNum.IsNumber() {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("timezone", shared.TypeNumber, "timestamp"))
	}

	tsSeconds, err := tsNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	instant := time.Unix(tsSeconds, 0).In(time.Local)

	abbreviation, offsetSeconds := instant.Zone()

	isDaylightSaving := constants.NUM_FAL

	if instant.IsDST() {
		isDaylightSaving = constants.NUM_TRU
	}

	keys := []string{
		"offsetSeconds",
		"abbreviation",
		"isDaylightSaving",
	}

	entries := map[string]values.Value{
		"offsetSeconds":    values.NewNumber(offsetSeconds),
		"abbreviation":     values.NewString(abbreviation),
		"isDaylightSaving": values.NewNumber(isDaylightSaving),
	}

	result := values.NewMapFromEntries(keys, entries)

	return res.Success(result)
}
