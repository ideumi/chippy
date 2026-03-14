/*
 *
 * RR2 - internal/builtins/loadopt.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/optional"
	"chip-go/internal/values"

	// Import optionals to trigger their init()
	_ "chip-go/internal/optional/hash"
	_ "chip-go/internal/optional/http"
	_ "chip-go/internal/optional/json"
	_ "chip-go/internal/optional/tls"
)

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

	// Look up the optional
	opt, exists := optional.GetOptional(optionalName.Value)

	if !exists {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Optional \""+optionalName.Value+"\" does not exist",
			ctx,
		))
	}

	// Get global context
	roadRunner := shared.GetGlobalRoadRunner2()

	if roadRunner == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"RoadRunner2 not available",
			ctx,
		))
	}

	globalCtx := roadRunner.GetGlobalContext()

	// Register funcs into GST
	for name, fn := range opt.Functions {
		// Silently succeed if already loaded
		existing := globalCtx.SymbolTable.Get(name)

		if existing == nil {
			fn.SetContext(globalCtx)

			globalCtx.SymbolTable.Set(name, fn)
		}
	}

	// Register constants into GST
	for name, constant := range opt.Constants {
		// Silently succeed if already loaded
		existing := globalCtx.SymbolTable.Get(name)

		if existing == nil {
			constant.SetContext(globalCtx)

			globalCtx.SymbolTable.Set(name, constant)
		}
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
