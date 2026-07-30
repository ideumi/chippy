/*
 *
 * RR2 - internal/builtins/args.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

var globalArgs []string

func SetGlobalArgs(arguments []string) {
	globalArgs = arguments
}

func argsFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("args", 0))
	}

	// To String
	elements := make([]values.Value, len(globalArgs))

	for i, arg := range globalArgs {
		elements[i] = values.NewString(arg)
	}

	return res.Success(values.NewList(elements))
}
