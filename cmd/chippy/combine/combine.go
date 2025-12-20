/*
 *
 * The Chippy Combine Tool
 *
 */

package combine

import (
	"chip-go/cmd/chippy/combine/safety"
	"chip-go/internal/constants"
	"chip-go/internal/roadrunner"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func HandleCombineCommand(args []string) {
	if len(args) > 0 && args[0] == "new" {
		if err := generateTemplateCombineFile(); err != nil {
			fmt.Printf("Error: %v\n", err)

			os.Exit(1)
		}
		return
	}

	combineFile := constants.COMBINE_DEFAULT_FILENAME

	if len(args) > 0 {
		combineFile = args[0]
	}

	if _, err := os.Stat(combineFile); os.IsNotExist(err) {
		fmt.Printf("Error: combine file '%s' not found\n", combineFile)

		os.Exit(1)
	}

	if err := executeCombine(combineFile); err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
}

func generateTemplateCombineFile() error {
	if _, err := os.Stat(constants.COMBINE_DEFAULT_FILENAME); err == nil {
		return fmt.Errorf("file '%s' already exists, rejected", constants.COMBINE_DEFAULT_FILENAME)
	}

	template := constants.COMBINE_SHEBANG + constants.COMBINE_TEMPLATE

	err := os.WriteFile(constants.COMBINE_DEFAULT_FILENAME, []byte(template), constants.FILE_PERM_READABLE)

	if err != nil {
		return fmt.Errorf("creating template file: %w", err)
	}

	fmt.Printf("Created %s\n", constants.COMBINE_DEFAULT_FILENAME)
	return nil
}

func executeCombine(combineFile string) error {
	// Read and execute combine file
	content, err := os.ReadFile(combineFile)

	if err != nil {
		return fmt.Errorf("reading combine file '%s': %w", combineFile, err)
	}

	rr := roadrunner.NewRoadRunner2()

	// Run combine.chp
	_, err = rr.Run(combineFile, string(content))

	if err != nil {
		return fmt.Errorf("executing combine file '%s': %w", combineFile, err)
	}

	// Extract configuration from context
	config := extractCombineConfig(rr.GetGlobalContext())

	if config.Project == "" {
		return fmt.Errorf("project name not specified in '%s'", combineFile)
	}

	if config.Output == "" {
		config.Output = config.Project
	}

	if config.Source == "" {
		return fmt.Errorf("source entry point not specified in '%s'", combineFile)
	}

	// Validate source file exists
	if _, err := os.Stat(config.Source); os.IsNotExist(err) {
		return fmt.Errorf("source file '%s' not found", config.Source)
	}

	// Build dependency graph from entry point
	deps, skipped, err := buildDependencyGraph(filepath.Clean(config.Source), config.Paths, config.External)

	if err != nil {
		return err
	}

	// Validate files for syntax errors and collisions
	validation, err := safety.ValidateFiles(deps)

	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Report parse errors
	if len(validation.Errors) > 0 {
		fmt.Println("Errors detected:")

		for _, verr := range validation.Errors {
			fmt.Printf("%v\n", verr.Error)
		}

		return fmt.Errorf("cannot bundle due to errors")
	}

	// Report symbol collisions
	if len(validation.Collisions) > 0 {
		reportCollisions(validation.Collisions)

		return fmt.Errorf("cannot bundle due to symbol collisions")
	}

	// Collect plugins and assets
	plugins, skippedPlugins, err := collectPlugins(deps, config.PluginPaths, config.ExternalPlugins)

	if err != nil {
		return fmt.Errorf("collecting plugins: %w", err)
	}

	// Check early if text bundle mode is incompatible with plugins (non-external plugins)
	if !config.ProduceBundle && len(plugins) > 0 {
		return fmt.Errorf("cannot bundle plugins with ProduceBundle = false (use ExternalPlugins to keep pload() calls)")
	}

	assets, err := collectAssets(config.Assets)

	if err != nil {
		return fmt.Errorf("collecting assets: %w", err)
	}

	// Show bundling information
	printBundleInfo(config, deps, plugins, assets, skipped, skippedPlugins)

	// Generate combined output
	if err := generateCombinedFile(config, deps, plugins, assets); err != nil {
		return err
	}

	fmt.Printf("Successfully created %s\n", config.Output)
	return nil
}

func reportCollisions(collisions []safety.SymbolCollision) {
	fmt.Println("Symbol collisions detected:")

	for _, collision := range collisions {
		if collision.IsBuiltinCollision() {
			fmt.Printf("\n  %s '%s' collides with %s:\n", collision.Locations[0].SymType, collision.Name, collision.SymType)

			for _, loc := range collision.Locations {
				fmt.Printf("    - %s:%d\n", loc.File, loc.Line+1)
			}
		} else if collision.IsSameFile() {
			fmt.Printf("\n  %s '%s' defined multiple times in %s:\n", collision.SymType, collision.Name, collision.GetFirstFile())

			for _, loc := range collision.Locations {
				fmt.Printf("    - Line %d\n", loc.Line+1)
			}
		} else {
			fmt.Printf("\n  %s '%s' defined in multiple files:\n", collision.SymType, collision.Name)

			for _, loc := range collision.Locations {
				fmt.Printf("    - %s:%d\n", loc.File, loc.Line+1)
			}
		}
	}
}

func printBundleInfo(config CombineConfig, deps []string, plugins []string, assets []string, skipped []string, skippedPlugins []string) {
	fmt.Printf("Bundling %s V-%s:\n", config.Project, config.Version)
	fmt.Printf("Output: %s\n", config.Output)
	fmt.Printf("\nFiles bundled:\n")

	for i, file := range deps {
		fmt.Printf("  %d. %s\n", i+1, file)
	}

	if len(plugins) > 0 {
		fmt.Printf("\nPlugins bundled:\n")

		for i, plugin := range plugins {
			fmt.Printf("  %d. %s\n", i+1, plugin)
		}
	}

	if len(assets) > 0 {
		fmt.Printf("\nAssets bundled:\n")

		for i, asset := range assets {
			fmt.Printf("  %d. %s\n", i+1, asset)
		}
	}

	if len(skipped) > 0 {
		fmt.Printf("\nExternal dependencies skipped:\n")

		sortedSkipped := make([]string, len(skipped))

		copy(sortedSkipped, skipped)

		sort.Strings(sortedSkipped)

		for i, file := range sortedSkipped {
			fmt.Printf("  %d. %s\n", i+1, file)
		}
	}

	if len(skippedPlugins) > 0 {
		fmt.Printf("\nExternal plugins skipped:\n")

		for i, plugin := range skippedPlugins {
			fmt.Printf("  %d. %s\n", i+1, plugin)
		}
	}

	// Print total
	totalMsg := fmt.Sprintf("\nTotal: %d source(s)", len(deps))

	if len(plugins) > 0 {
		totalMsg += fmt.Sprintf(", %d plugin(s)", len(plugins))
	}

	if len(assets) > 0 {
		totalMsg += fmt.Sprintf(", %d asset(s)", len(assets))
	}

	if len(skipped) > 0 {
		totalMsg += fmt.Sprintf(", %d external", len(skipped))
	}

	if len(skippedPlugins) > 0 {
		totalMsg += fmt.Sprintf(", %d external plugin(s)", len(skippedPlugins))
	}

	fmt.Println(totalMsg)
}
