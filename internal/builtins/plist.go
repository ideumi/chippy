/*
 *
 * RR2 - internal/builtins/plist.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
)

func plistFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("plist", 0),
			ctx,
		))
	}

	plugins := shared.GetAllPlugins()

	// [[name, version, path, [functions...], [constants...]], ...]
	var resultElements []values.Value

	for _, plugin := range plugins {
		// Functions
		var funcElements []values.Value

		for _, funcName := range plugin.Functions {
			funcElements = append(funcElements, values.NewString(funcName).SetContext(ctx))
		}

		funcList := values.NewList(funcElements).SetContext(ctx)

		// Constants
		var constElements []values.Value

		for _, constName := range plugin.Constants {
			constElements = append(constElements, values.NewString(constName).SetContext(ctx))
		}

		constList := values.NewList(constElements).SetContext(ctx)

		pluginInfo := values.NewList([]values.Value{
			values.NewString(plugin.Name).SetContext(ctx),
			values.NewString(plugin.Version).SetContext(ctx),
			values.NewString(plugin.Path).SetContext(ctx),
			funcList,
			constList,
		}).SetContext(ctx)

		resultElements = append(resultElements, pluginInfo)
	}

	result := values.NewList(resultElements)
	return res.Success(result.SetContext(ctx))
}
