/*
 *
 * The Chippy Combine Tool - Configuration
 *
 */

package combine

import (
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/values"
)

type CombineConfig struct {
	Project         string
	Version         string
	Licence         string
	Output          string
	Source          string
	Paths           []string
	PluginPaths     []string
	Assets          []string
	External        []string
	ExternalPlugins []string
	StripComments   bool
	StripWhitespace bool
	ProduceBundle   bool
}

func extractCombineConfig(ctx *context.Context) CombineConfig {
	config := CombineConfig{}

	// Extract string variables
	if val := ctx.SymbolTable.Get(constants.CONFIG_PROJECT); val != nil {
		if str, ok := val.(*values.String); ok {
			config.Project = str.Value
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_VERSION); val != nil {
		if str, ok := val.(*values.String); ok {
			config.Version = str.Value
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_LICENCE); val != nil {
		if str, ok := val.(*values.String); ok {
			config.Licence = str.Value
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_OUTPUT); val != nil {
		if str, ok := val.(*values.String); ok {
			config.Output = str.Value
		}
	}

	// Extract source entry point
	if val := ctx.SymbolTable.Get(constants.CONFIG_SOURCE); val != nil {
		if str, ok := val.(*values.String); ok {
			config.Source = str.Value
		}
	}

	// Extract search paths
	if val := ctx.SymbolTable.Get(constants.CONFIG_PATHS); val != nil {
		if list, ok := val.(*values.List); ok {
			for _, elem := range list.Elements {
				if str, ok := elem.(*values.String); ok {
					config.Paths = append(config.Paths, str.Value)
				}
			}
		}
	}

	// Extract plugin paths
	if val := ctx.SymbolTable.Get(constants.CONFIG_PLUGIN_PATHS); val != nil {
		if list, ok := val.(*values.List); ok {
			for _, elem := range list.Elements {
				if str, ok := elem.(*values.String); ok {
					config.PluginPaths = append(config.PluginPaths, str.Value)
				}
			}
		}
	}

	// Extract assets
	if val := ctx.SymbolTable.Get(constants.CONFIG_ASSETS); val != nil {
		if list, ok := val.(*values.List); ok {
			for _, elem := range list.Elements {
				if str, ok := elem.(*values.String); ok {
					config.Assets = append(config.Assets, str.Value)
				}
			}
		}
	}

	// Extract external dependencies
	if val := ctx.SymbolTable.Get(constants.CONFIG_EXTERNAL); val != nil {
		if list, ok := val.(*values.List); ok {
			for _, elem := range list.Elements {
				if str, ok := elem.(*values.String); ok {
					config.External = append(config.External, str.Value)
				}
			}
		}
	}

	// Extract external plugins
	if val := ctx.SymbolTable.Get(constants.CONFIG_EXTERNAL_PLUGINS); val != nil {
		if list, ok := val.(*values.List); ok {
			for _, elem := range list.Elements {
				if str, ok := elem.(*values.String); ok {
					config.ExternalPlugins = append(config.ExternalPlugins, str.Value)
				}
			}
		}
	}

	// Flags
	if val := ctx.SymbolTable.Get(constants.CONFIG_STRIP_COMMENTS); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.StripComments = num.Value != constants.NUM_NUL
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_STRIP_WHITESPACE); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.StripWhitespace = num.Value != constants.NUM_NUL
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_PRODUCE_BUNDLE); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.ProduceBundle = num.Value != constants.NUM_NUL
		}
	}

	return config
}
