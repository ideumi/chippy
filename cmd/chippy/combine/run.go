/*
 *
 * Combine Run Command
 *
 */

package combine

import (
	"chip-go/internal/builtins"
	"chip-go/internal/constants"
	"chip-go/internal/modena"
	"chip-go/internal/values"
	"errors"
	"fmt"
	"os"
)

var (
	errRunErrors     = errors.New(constants.COMBINE_E_RUN_ERRORS)
	errRunCollisions = errors.New(constants.COMBINE_E_RUN_COLLISIONS)
)

type resolvedFile struct {
	path   string
	source string
}

type resolvedProgram struct {
	deps       []resolvedFile
	entry      resolvedFile
	parseErrs  []ValidationError
	collisions []SymbolCollision
}

func Run(cfgFile string, args []string) error {
	config, err := loadCombineConfig(cfgFile)

	if err != nil {
		return err
	}

	if err := validateCombineSource(config, cfgFile); err != nil {
		return err
	}

	program, err := resolveProgram(config.Source, config.Paths, config.External)

	if err != nil {
		return err
	}

	if len(program.parseErrs) > 0 {
		reportParseErrors(program.parseErrs)

		return errRunErrors
	}

	if len(program.collisions) > 0 {
		reportCollisions(program.collisions)

		return errRunCollisions
	}

	// Important API: pass args after chippy combine [config] run [args]
	value, halted, err := executeResolved(program, modena.New(), args)

	if err != nil {
		fmt.Println(err.Error())

		os.Exit(1)
	}

	if halted {
		exitWithResult(value)
	}

	return nil
}

func resolveProgram(source string, paths []string, external []string) (*resolvedProgram, error) {
	files, _, err := buildDependencyGraph(source, paths, external)

	if err != nil {
		return nil, err
	}

	validation, err := ValidateFiles(files)

	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	externalMap := make(map[string]bool)

	for _, ext := range external {
		externalMap[ext] = true
	}

	program := &resolvedProgram{
		parseErrs:  validation.Errors,
		collisions: validation.Collisions,
	}

	for i, path := range files {
		data, err := os.ReadFile(path)

		if err != nil {
			return nil, fmt.Errorf("reading file '%s': %w", path, err)
		}

		file := resolvedFile{path: path, source: removeLoadLines(string(data), externalMap)}

		if i == len(files)-1 {
			program.entry = file
		} else {
			program.deps = append(program.deps, file)
		}
	}

	return program, nil
}

func executeResolved(program *resolvedProgram, mod *modena.Modena, progArgs []string) (values.Value, bool, error) {
	// Set provided args
	builtins.SetGlobalArgs(progArgs)

	for _, dep := range program.deps {
		value, halted, err := mod.Run(dep.path, dep.source)

		if err != nil {
			return value, false, err
		}

		if halted {
			return value, true, nil
		}
	}

	value, halted, err := mod.Run(program.entry.path, program.entry.source)

	return value, halted, err
}
