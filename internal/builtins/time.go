/*
 *
 * RR2 - internal/builtins/time.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"time"
)

func timeFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("time", 0))
	}

	// Return high-precision timestamp
	timestamp := float64(time.Now().UnixNano()) / 1e9

	num, err := values.NewNumberFromFloat(timestamp)

	if err != nil {
		return res.Fail(err.Error())
	}

	return res.Success(num)
}
