/*
 *
 * RR2 - internal/builtins/args.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

var globalArgs []string

func SetGlobalArgs(arguments []string) {
	globalArgs = arguments
}

func argsFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("args", 0),
			ctx,
		))
	}

	// To String
	elements := make([]values.Value, len(globalArgs))

	for i, arg := range globalArgs {
		elements[i] = values.NewString(arg).SetContext(ctx)
	}

	return res.Success(values.NewList(elements).SetContext(ctx))
}
