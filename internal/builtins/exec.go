/*
 *
 * RR2 - internal/builtins/exec.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"syscall"
)

func execFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("exec", 1, "args"),
			ctx,
		))
	}

	argsList, ok := args[0].(*values.List)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("exec", shared.TypeList, "args"),
			ctx,
		))
	}

	var execArgs []string

	for _, elem := range argsList.Elements {
		str, ok := elem.(*values.String)
		if !ok {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("All arguments must be strings"),
				ctx,
			))
		}
		execArgs = append(execArgs, str.Value)
	}

	if len(execArgs) == 0 {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Arguments list cannot be empty"),
			ctx,
		))
	}

	// Use execv syscall to replace current process
	program := execArgs[0]
	err := syscall.Exec(program, execArgs, nil)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
