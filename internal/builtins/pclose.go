/*
 *
 * RR2 - internal/builtins/pclose.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os/exec"
)

func pcloseFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("pclose", 1, "handle"))
	}

	handleNum := args[0]

	if !handleNum.IsNumber() {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("pclose", shared.TypeNumber, "handle"))
	}

	handle64, err := handleNum.AsInt()

	if err != nil {
		return res.Failure(err)
	}

	handle := int(handle64)

	// Check for standard handles (cannot close these)
	if handle >= 0 && handle <= 2 {
		return res.FailAt(1, shared.Errors.InvalidValue("Cannot close standard handles (0, 1, 2)"))
	}

	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	procHandle, exists := registry.Processes.Extract(handle)

	if !exists {
		return res.FailAt(1, shared.Errors.InvalidValue("Invalid process handle"))
	}

	registry.Alloc.Free(handle)

	// Close pipes if they exist
	if procHandle.Stdin != nil {
		procHandle.Stdin.Close()
	}

	if procHandle.Stdout != nil {
		procHandle.Stdout.Close()
	}

	if procHandle.Stderr != nil {
		procHandle.Stderr.Close()
	}

	// Wait for process to finish and get exit code
	err = procHandle.Cmd.Wait()
	exitCode := 0

	if err != nil {
		// Check if non-zero exit
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Some other error (process could not run etc.)
			return res.Success(values.NewString(constants.STR_ERR))
		}
	}

	return res.Success(values.NewNumber(exitCode))
}
