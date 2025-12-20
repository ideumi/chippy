/*
 *
 * RR2 - internal/builtins/pload.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"fmt"
	"plugin"
)

func cleanupPluginRegistration(globalCtx *context.Context, registeredSymbols []string) {
	for _, name := range registeredSymbols {
		globalCtx.SymbolTable.Remove(name)
	}
}

func ploadFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("pload", 1, "filename"),
			ctx,
		))
	}

	filenameVal, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("pload", shared.TypeString, "filename"),
			ctx,
		))
	}

	// Open the plugin file
	p, err := plugin.Open(filenameVal.Value)

	if err != nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Failed to pload: "+err.Error(),
			ctx,
		))
	}

	// Look up GetPlugin function
	sym, err := p.Lookup("GetPlugin")

	if err != nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Plugin missing GetPlugin() function",
			ctx,
		))
	}

	getPlugin, ok := sym.(func() (string, string, string, map[string]*values.NativeFunction, map[string]values.Value))

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Invalid GetPlugin() signature",
			ctx,
		))
	}

	pluginName, pluginVersion, pluginChipLangVersion, builtinFuncs, consts := getPlugin()

	// Check if already loaded
	if _, loaded := shared.GetPlugin(pluginName); loaded {
		// Already loaded, return success
		return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
	}

	// Verify version
	if pluginChipLangVersion != constants.STR_LPLVR {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"\""+pluginName+"\" V-"+pluginVersion+" requires V-"+pluginChipLangVersion+", running V-"+constants.STR_LPLVR,
			ctx,
		))
	}

	rr, ok := globalRoadRunner2.(interface{ GetGlobalContext() *context.Context })

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Cannot access global context, something is seriously wrong",
			ctx,
		))
	}

	globalCtx := rr.GetGlobalContext()

	// Check for collissions within the plugin itself
	seenNames := make(map[string]bool)

	for name := range builtinFuncs {
		if seenNames[name] {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"\""+pluginName+"\" exports duplicate symbol \""+name+"\"",
				ctx,
			))
		}
		seenNames[name] = true
	}

	for name := range consts {
		if seenNames[name] {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"\""+pluginName+"\" exports duplicate symbol \""+name+"\"",
				ctx,
			))
		}
		seenNames[name] = true
	}

	var functionNames []string
	var constantNames []string
	var registeredSymbols []string

	// _pluginname_functionname_
	for name, fn := range builtinFuncs {
		if !fn.Plugin {
			posStart, posEnd := args[0].GetPos()

			cleanupPluginRegistration(globalCtx, registeredSymbols)

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Function \""+name+"\" not marked as plugin function",
				ctx,
			))
		}

		prefixedName := fmt.Sprintf("_%s_%s_", pluginName, name)

		// Collission detection
		existing := globalCtx.SymbolTable.Get(prefixedName)

		if existing != nil {
			posStart, posEnd := args[0].GetPos()

			cleanupPluginRegistration(globalCtx, registeredSymbols)

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Function \""+prefixedName+"\" already exists",
				ctx,
			))
		}

		fn.SetContext(globalCtx)
		globalCtx.SymbolTable.Set(prefixedName, fn)

		functionNames = append(functionNames, prefixedName)
		registeredSymbols = append(registeredSymbols, prefixedName)
	}

	// Same for constants
	for name, val := range consts {
		prefixedName := fmt.Sprintf("_%s_%s_", pluginName, name)

		existing := globalCtx.SymbolTable.Get(prefixedName)

		if existing != nil {
			posStart, posEnd := args[0].GetPos()

			cleanupPluginRegistration(globalCtx, registeredSymbols)

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				"Constant \""+prefixedName+"\" already exists",
				ctx,
			))
		}

		val.SetContext(globalCtx)
		globalCtx.SymbolTable.Set(prefixedName, val)
		constantNames = append(constantNames, prefixedName)
		registeredSymbols = append(registeredSymbols, prefixedName)
	}

	// Register plugin in registry
	shared.RegisterPlugin(&shared.LoadedPlugin{
		Name:      pluginName,
		Version:   pluginVersion,
		Path:      filenameVal.Value,
		Functions: functionNames,
		Constants: constantNames,
	})

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
