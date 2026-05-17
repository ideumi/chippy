/*
 *
 * RR2 - internal/builtins/spawn.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
	"syscall"
)

func spawnFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("spawn", 1, "args")))
	}

	argsList, ok := args[0].(*values.List)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("spawn", shared.TypeList, "args")))
	}

	var spawnArgs []string

	for _, elem := range argsList.Elements {
		str, ok := elem.(*values.String)

		if !ok {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("All arguments must be strings")))
		}

		spawnArgs = append(spawnArgs, str.Value)
	}

	if len(spawnArgs) == 0 {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidValue("Arguments list cannot be empty")))
	}

	attr := &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{nil, nil, nil},
		Sys:   &syscall.SysProcAttr{Setsid: true},
	}

	proc, err := os.StartProcess(spawnArgs[0], spawnArgs, attr)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	pid := proc.Pid

	// Release the process so it doesn't become a zombie
	go proc.Wait()

	return res.Success(values.NewNumber(pid).SetContext(ctx))
}
