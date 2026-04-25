/*
 *
 * RR2 - internal/builtins/loadopt.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
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
func InstallOpt(opt *optional.Optional, globalCtx *context.Context) {
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

func loadoptFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("loadopt", 1, "optional"),
			ctx,
		))
	}

	optionalName, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("loadopt", shared.TypeString, "optional"),
			ctx,
		))
	}

	opt, exists := optional.GetOptional(optionalName.Value)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Optional \""+optionalName.Value+"\" does not exist",
			ctx,
		))
	}

	roadRunner := orchestrator.Get().GetRR2ForContext(ctx)

	if roadRunner == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"RoadRunner2 not available",
			ctx,
		))
	}

	globalCtx := roadRunner.GetGlobalContext()

	InstallOpt(opt, globalCtx)

	instanceID := context.GetInstanceID(ctx)
	orch := orchestrator.Get()

	if inst := orch.GetInstance(instanceID); inst != nil {
		orch.AddLoadedOpt(inst, optionalName.Value)
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
