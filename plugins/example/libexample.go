/*
 *
 * Structural Reference Plugin
 *
 */

package main

import (
	"chip-go/internal/values"
)

const (
	PluginName      = "example"
	PluginVersion   = "1.0.0"
	ChipLangVersion = "1.0.6"
)

func GetPlugin() (string, string, string, map[string]*values.NativeFunction, map[string]values.Value) {
	pluginFunctions := make(map[string]*values.NativeFunction)
	pluginConstants := make(map[string]values.Value)

	pluginFunctions["test"] = values.NewNativeFunction("test", testFunc, values.Plugin)

	pluginConstants["TESTCONST"] = values.NewNumber(32)

	return PluginName, PluginVersion, ChipLangVersion, pluginFunctions, pluginConstants
}

func testFunc(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	return res.Success(values.NewString("Hello from plugin!").SetContext(ctx))
}
