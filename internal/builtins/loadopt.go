/*
 *
 * RR2 - internal/builtins/loadopt.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/optional"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"

	// Import optionals to trigger their init()
	_ "chip-go/internal/optional/hash"
	_ "chip-go/internal/optional/http"
	_ "chip-go/internal/optional/json"
	_ "chip-go/internal/optional/tls"
)

// InstallOpt registers an optional's functions and constants into the given GST.
// Names already present are left alone so a second loadopt(optional) of the same
// optional is a silent no-op.
//
// Used both by loadopt(optional) and by actor spawn to inherit the spawner's
// loaded optionals.
func InstallOpt(opt *optional.Optional, globalCtx values.Ctx) {
	for name, fn := range opt.Functions {
		if existing := globalCtx.SymbolTable.Get(name); existing == nil {
			fn.SetContext(globalCtx)
			globalCtx.SymbolTable.Set(name, fn)
		}
	}

	for name, constant := range opt.Constants {
		if existing := globalCtx.SymbolTable.Get(name); existing == nil {
			constant.SetContext(globalCtx)
			globalCtx.SymbolTable.Set(name, constant)
		}
	}
}

func loadoptFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("loadopt", 1, "optional"))
	}

	optionalName, ok := args[0].(*values.String)

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("loadopt", shared.TypeString, "optional"))
	}

	opt, exists := optional.GetOptional(optionalName.Value)

	if !exists {
		return res.FailAt(1, "Optional \""+optionalName.Value+"\" does not exist")
	}

	mod := orchestrator.Get().GetModenaForContext(ctx.InstanceID)

	if mod == nil {
		return res.FailAt(1, "Modena not available")
	}

	globalCtx := mod.GetGlobalContext()

	InstallOpt(opt, globalCtx)

	instanceID := ctx.InstanceID
	orch := orchestrator.Get()

	if inst := orch.GetInstance(instanceID); inst != nil {
		orch.AddLoadedOpt(inst, optionalName.Value)
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
