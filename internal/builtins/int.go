/*
 *
 * RR2 - internal/builtins/int.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
)

func intFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("int", 1, "value"))
	}

	num := args[0]
	if !num.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("int", shared.TypeNumber, "value"))
	}

	intVal, err := num.AsInt()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewNumber(intVal))
}
