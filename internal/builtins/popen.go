/*
 *
 * RR2 - internal/builtins/popen.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/handles"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os/exec"
	"strings"
)

func popenFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("popen", 2, "command, mode")))
	}

	commandStr, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("popen", shared.PositionFirst, shared.TypeString, "command")))
	}

	modeStr, ok := args[1].(*values.String)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("popen", shared.PositionSecond, shared.TypeString, "mode")))
	}

	command := commandStr.Value
	mode := modeStr.Value

	// Use /bin/sh as per POSIX standard and Python's approach
	// This avoids PATH dependencies in forked processes
	cmd := exec.Command("/bin/sh", "-c", command)

	var procHandle handles.ProcessHandle
	procHandle.Cmd = cmd

	var err error

	// Set up pipes based on mode
	if strings.Contains(mode, "w") || mode == "rw" {
		procHandle.Stdin, err = cmd.StdinPipe()

		if err != nil {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Failed to create stdin pipe: "+err.Error()))
		}
	}

	if strings.Contains(mode, "r") || mode == "rw" {
		procHandle.Stdout, err = cmd.StdoutPipe()

		if err != nil {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Failed to create stdout pipe: "+err.Error()))
		}
	}

	// Start the process
	err = cmd.Start()

	if err != nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Failed to start process: "+err.Error()))
	}

	// Allocate handle using recycling system
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	handle := registry.Alloc.Alloc()
	registry.Processes.Store(handle, &procHandle)

	return res.Success(values.NewNumber(handle).SetContext(ctx))
}
