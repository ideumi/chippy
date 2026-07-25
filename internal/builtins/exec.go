/*
 *
 * RR2 - internal/builtins/exec.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
	"syscall"
)

func execFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("exec", 1, "args"))
	}

	argsList, ok := args[0].(*values.List)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("exec", shared.TypeList, "args"))
	}

	var execArgs []string

	for _, elem := range argsList.Elements {
		str, ok := elem.(*values.String)
		if !ok {
			return res.FailAt(1, shared.Errors.InvalidValue("All arguments must be strings"))
		}
		execArgs = append(execArgs, str.Value)
	}

	if len(execArgs) == 0 {
		return res.FailAt(1, shared.Errors.InvalidValue("Arguments list cannot be empty"))
	}

	program := execArgs[0]
	err := syscall.Exec(program, execArgs, os.Environ())

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
