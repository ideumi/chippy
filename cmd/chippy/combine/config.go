/*
 *
 * Combine Config Loader
 *
 */

package combine

import (
	"chip-go/internal/constants"
	"chip-go/internal/modena"
	"chip-go/internal/values"
	"fmt"
	"os"
)

type CombineConfig struct {
	Project         string
	Version         string
	Licence         string
	Output          string
	Source          string
	Paths           []string
	External        []string
	StripComments   bool
	StripWhitespace bool
	AddShebang      bool
	Compile         bool
}

func runCombineConfigFile(cfgFile string) (values.Ctx, error) {
	content, err := os.ReadFile(cfgFile)

	if err != nil {
		return nil, fmt.Errorf("reading combine file '%s': %w", cfgFile, err)
	}

	mod := modena.New()

	value, halted, err := mod.Run(cfgFile, string(content))

	if err != nil {
		return nil, fmt.Errorf("executing combine file '%s': %w", cfgFile, err)
	}

	if halted {
		exitWithResult(value)
	}

	return mod.GetGlobalContext(), nil
}

// loadCombineConfig runs the combine config in a new VM and read its globals
func loadCombineConfig(cfgFile string) (CombineConfig, error) {
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return CombineConfig{}, fmt.Errorf("combine file '%s' not found", cfgFile)
	}

	ctx, err := runCombineConfigFile(cfgFile)

	if err != nil {
		return CombineConfig{}, err
	}

	return extractCombineConfig(ctx), nil
}

func extractCombineConfig(ctx values.Ctx) CombineConfig {
	config := CombineConfig{}

	// Extract string variables
	if val := ctx.Globals.GetByName(constants.CONFIG_PROJECT); val.IsSet() {
		if str, ok := values.AsString(val); ok {
			config.Project = str.Value
		}
	}

	if val := ctx.Globals.GetByName(constants.CONFIG_VERSION); val.IsSet() {
		if str, ok := values.AsString(val); ok {
			config.Version = str.Value
		}
	}

	if val := ctx.Globals.GetByName(constants.CONFIG_LICENCE); val.IsSet() {
		if str, ok := values.AsString(val); ok {
			config.Licence = str.Value
		}
	}

	if val := ctx.Globals.GetByName(constants.CONFIG_OUTPUT); val.IsSet() {
		if str, ok := values.AsString(val); ok {
			config.Output = str.Value
		}
	}

	// Extract source entry point
	if val := ctx.Globals.GetByName(constants.CONFIG_SOURCE); val.IsSet() {
		if str, ok := values.AsString(val); ok {
			config.Source = str.Value
		}
	}

	// Extract search paths
	if val := ctx.Globals.GetByName(constants.CONFIG_PATHS); val.IsSet() {
		if list, ok := values.AsList(val); ok {
			for _, elem := range list.Elements {
				if str, ok := values.AsString(elem); ok {
					config.Paths = append(config.Paths, str.Value)
				}
			}
		}
	}

	// Extract external dependencies
	if val := ctx.Globals.GetByName(constants.CONFIG_EXTERNAL); val.IsSet() {
		if list, ok := values.AsList(val); ok {
			for _, elem := range list.Elements {
				if str, ok := values.AsString(elem); ok {
					config.External = append(config.External, str.Value)
				}
			}
		}
	}

	// Flags
	if val := ctx.Globals.GetByName(constants.CONFIG_STRIP_COMMENTS); val.IsNumber() {
		config.StripComments = val.IsTrue()
	}

	if val := ctx.Globals.GetByName(constants.CONFIG_STRIP_WHITESPACE); val.IsNumber() {
		config.StripWhitespace = val.IsTrue()
	}

	if val := ctx.Globals.GetByName(constants.CONFIG_ADD_SHEBANG); val.IsNumber() {
		config.AddShebang = val.IsTrue()
	}

	if val := ctx.Globals.GetByName(constants.CONFIG_COMPILE); val.IsNumber() {
		config.Compile = val.IsTrue()
	}

	return config
}

func validateCombineSource(config CombineConfig, cfgFile string) error {
	if config.Source == "" {
		return fmt.Errorf("source entry point not specified in '%s'", cfgFile)
	}

	if _, err := os.Stat(config.Source); os.IsNotExist(err) {
		return fmt.Errorf("source file '%s' not found", config.Source)
	}

	return nil
}
