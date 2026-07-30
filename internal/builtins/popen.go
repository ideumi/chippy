/*
 *
 * RR2 - internal/builtins/popen.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/handles"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os/exec"
	"strings"
)

func popenFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("popen", 2, "command, mode"))
	}

	commandStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypePositionalWithHint("popen", shared.PositionFirst, shared.TypeString, "command"))
	}

	modeStr, ok := values.AsString(args[1])

	if !ok {
		return res.FailAt(2,
			shared.Errors.InvalidArgTypePositionalWithHint("popen", shared.PositionSecond, shared.TypeString, "mode"))
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
			return res.FailAt(1, "Failed to create stdin pipe: "+err.Error())
		}
	}

	if strings.Contains(mode, "r") || mode == "rw" {
		procHandle.Stdout, err = cmd.StdoutPipe()

		if err != nil {
			return res.FailAt(1, "Failed to create stdout pipe: "+err.Error())
		}
	}

	// Start the process
	err = cmd.Start()

	if err != nil {
		return res.FailAt(1, "Failed to start process: "+err.Error())
	}

	// Allocate handle using recycling system
	registry := orchestrator.Get().GetRegistry(ctx.InstanceID)
	handle := registry.Alloc.Alloc()
	registry.Processes.Store(handle, &procHandle)

	return res.Success(values.NewNumber(handle))
}
