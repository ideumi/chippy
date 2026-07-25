/*
 *
 * RR2 - internal/builtins/sin.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"math"
)

func sinFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("sin", 1, "radians"))
	}

	radiansArg, ok := args[0].(*values.Number)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("sin", shared.TypeNumber, "radians"))
	}

	radians := radiansArg.AsFloat()

	num, err := values.NewNumberFromFloat(math.Sin(radians))

	if err != nil {
		return res.FailAt(1, err.Error())
	}

	return res.Success(num.SetContext(ctx))
}
