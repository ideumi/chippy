/*
 *
 * Chippy - internal/builtins/spawn.go
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

func spawnFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("spawn", 1, "args"))
	}

	argsList, ok := values.AsList(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("spawn", shared.TypeList, "args"))
	}

	var spawnArgs []string

	for _, elem := range argsList.Elements {
		str, ok := values.AsString(elem)

		if !ok {
			return res.FailAt(1, shared.Errors.InvalidValue("All arguments must be strings"))
		}

		spawnArgs = append(spawnArgs, str.Value)
	}

	if len(spawnArgs) == 0 {
		return res.FailAt(1, shared.Errors.InvalidValue("Arguments list cannot be empty"))
	}

	attr := &os.ProcAttr{
		Env:   os.Environ(),
		Files: []*os.File{nil, nil, nil},
		Sys:   &syscall.SysProcAttr{Setsid: true},
	}

	proc, err := os.StartProcess(spawnArgs[0], spawnArgs, attr)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	pid := proc.Pid

	// Release the process so it doesn't become a zombie
	go proc.Wait()

	return res.Success(values.NewNumber(pid))
}
