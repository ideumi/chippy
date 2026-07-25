/*
 *
 * RR2 - internal/builtins/load.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os"
)

func loadFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("load", 1, "filename"))
	}

	filename, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("load", shared.TypeString, "filename"))
	}

	content, err := os.ReadFile(filename.Value)

	if err != nil {
		return res.FailAt(1, "Failed to load file \""+filename.Value+"\": "+err.Error())
	}

	mod := orchestrator.Get().GetModenaForContext(ctx.InstanceID)

	if mod == nil {
		return res.FailAt(1, "Modena not available")
	}

	_, err = mod.Run(filename.Value, string(content))

	if err != nil {
		return res.FailAt(1, "Failed to execute file \""+filename.Value+"\":\n"+err.Error())
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
