/*
 *
 * RR2 - internal/builtins/timezone.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"time"
)

func timezoneFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("timezone", 1, "timestamp")))
	}

	tsNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("timezone", shared.TypeNumber, "timestamp")))
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
		"offsetSeconds":    values.NewNumber(offsetSeconds).SetContext(ctx),
		"abbreviation":     values.NewString(abbreviation).SetContext(ctx),
		"isDaylightSaving": values.NewNumber(isDaylightSaving).SetContext(ctx),
	}

	result := values.NewMapFromEntries(keys, entries)

	return res.Success(result.SetContext(ctx))
}
