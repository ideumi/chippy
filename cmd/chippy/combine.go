/*
 *
 * The Chippy Combine Tool
 *
 */

package main

import (
	"bufio"
	"chip-go/cmd/chippy/safety"
	"chip-go/internal/constants"
	"chip-go/internal/roadrunner"
	"chip-go/internal/values"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// FIXME: This should probably be optimized, three file reads is not optimal, but oh well.
var (
	loadRegex     = regexp.MustCompile(`load\s*\(\s*"((?:[^"\\]|\\.)*)"\s*\)`)
	loadCallRegex = regexp.MustCompile(`^\s*load\s*\(`)
)

func reportCollisions(collisions []safety.SymbolCollision) {
	fmt.Println("Symbol collisions detected:")

	for _, collision := range collisions {
		// Handle builtin collisions differently

		if collision.IsBuiltinCollision() {
			fmt.Printf("\n  %s '%s' collides with %s:\n", collision.Locations[0].SymType, collision.Name, collision.SymType)

			for _, loc := range collision.Locations {
				fmt.Printf("    - %s:%d\n", loc.File, loc.Line+1)
			}
		} else if collision.IsSameFile() {
			fmt.Printf("\n  %s '%s' defined multiple times in %s:\n", collision.SymType, collision.Name, collision.GetFirstFile())

			for _, loc := range collision.Locations {
				fmt.Printf("    - %s:%d\n", loc.File, loc.Line+1)
			}
		} else {
			fmt.Printf("\n  %s '%s' defined in multiple files:\n", collision.SymType, collision.Name)

			for _, loc := range collision.Locations {
				fmt.Printf("    - %s:%d\n", loc.File, loc.Line+1)
			}
		}
	}
}

func handleCombineCommand(args []string) {
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
	// Check if combine.chp already exists
	if _, err := os.Stat(constants.COMBINE_DEFAULT_FILENAME); err == nil {
		return fmt.Errorf("file '%s' already exists, rejected", constants.COMBINE_DEFAULT_FILENAME)
	}

	template := constants.COMBINE_SHEBANG + `
#
# Combine configuration file
#

# Metadata

var Project = "myprogram";
var Version = "1.0.0";
var Licence = "licence.txt";
var Output = "myprogram";

# Entry point

var Source = "main.chp";

# Search paths for dependencies

var Paths = [];
Paths = Paths + ["src/lib"];

# Add user-local library path if HOME is available

var homePath = getenv("HOME");

if not iserr(homePath) {
	Paths = Paths + [homePath + "/.local/share/chippy/lib"];
}

Paths = Paths + ["/usr/lib/chippy"];

# Dependencies combine shouldn't bundle and just keep as loads

var External = [];

# Options

var StripComments = true;
var StripWhitespace = true;
var AddShebang = true;
`

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

	// Validate configuration
	if config.StripWhitespace && !config.StripComments {
		return fmt.Errorf("invalid configuration: StripWhitespace requires StripComments")
	}

	// Build dependency graph from entry point
	deps, skipped, err := buildDependencyGraph(config.Source, config.Paths, config.External)

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

	// Show bundling information
	fmt.Printf("Bundling %s V-%s:\n", config.Project, config.Version)
	fmt.Printf("Output: %s\n", config.Output)
	fmt.Printf("\nFiles bundled:\n")

	for i, file := range deps {
		fmt.Printf("  %d. %s\n", i+1, file)
	}

	if len(skipped) > 0 {
		fmt.Printf("\nExternal dependencies skipped:\n")

		for i, file := range skipped {
			fmt.Printf("  %d. %s\n", i+1, file)
		}
	}

	if len(skipped) > 0 {
		fmt.Printf("\nTotal: %d source(s), %d external\n", len(deps), len(skipped))
	} else {
		fmt.Printf("\nTotal: %d source(s)\n", len(deps))
	}

	// Generate combined output
	if err := generateCombinedFile(config, deps); err != nil {
		return err
	}

	fmt.Printf("Successfully created %s\n", config.Output)

	return nil
}

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
}

func extractCombineConfig(ctx values.Ctx) CombineConfig {
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

	// Flags
	if val := ctx.SymbolTable.Get(constants.CONFIG_STRIP_COMMENTS); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.StripComments = num.IsTrue()
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_STRIP_WHITESPACE); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.StripWhitespace = num.IsTrue()
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_ADD_SHEBANG); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.AddShebang = num.IsTrue()
		}
	}

	return config
}

func buildDependencyGraph(source string, paths []string, external []string) ([]string, []string, error) {
	seen := make(map[string]bool)
	skippedMap := make(map[string]bool)
	var result []string

	// Build external lookup map
	externalMap := make(map[string]bool)

	for _, ext := range external {
		externalMap[ext] = true
	}

	// Process dependencies starting from entry point
	if err := processDependencies(source, paths, externalMap, seen, skippedMap, &result); err != nil {
		return nil, nil, fmt.Errorf("processing dependencies: %w", err)
	}

	// Convert skipped map to slice
	var skipped []string

	for name := range skippedMap {
		skipped = append(skipped, name)
	}

	sort.Strings(skipped)

	return result, skipped, nil
}

func processDependencies(filename string, paths []string, external map[string]bool, seen map[string]bool, skipped map[string]bool, result *[]string) error {
	if seen[filename] {
		return nil
	}

	// Mark as seen to prevent cycles
	seen[filename] = true

	// Read and find load() calls
	content, err := os.ReadFile(filename)

	if err != nil {
		return fmt.Errorf("could not read %s: %w", filename, err)
	}

	// Extract load() calls
	matches := loadRegex.FindAllStringSubmatch(string(content), -1)

	// Process dependencies first
	for _, match := range matches {
		if len(match) > 1 {
			depPath := match[1]

			// Check if dependency is external
			if external[depPath] {
				skipped[depPath] = true

				continue
			}

			// Resolve dependency through search paths
			resolvedPath := resolvePath(depPath, paths)

			if resolvedPath == "" {
				return fmt.Errorf("dependency '%s' not found in search paths %v", depPath, paths)
			}

			if !seen[resolvedPath] {
				if err := processDependencies(resolvedPath, paths, external, seen, skipped, result); err != nil {
					return err
				}
			}
		}
	}

	// Add this file after its dependencies
	*result = append(*result, filename)

	return nil
}

func resolvePath(filename string, paths []string) string {
	// Clean the filename to normalize path separators and resolve . and .. elements
	filename = filepath.Clean(filename)

	// If file exists as-is (relative to current directory or absolute path), return it
	if _, err := os.Stat(filename); err == nil {
		return filename
	}

	// If no search paths specified, return empty
	if len(paths) == 0 {
		return ""
	}

	// Search in each path
	for _, path := range paths {
		fullPath := filepath.Clean(filepath.Join(path, filename))

		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return ""
}

func generateCombinedFile(config CombineConfig, files []string) error {
	var combined strings.Builder

	// Add shebang if requested
	if config.AddShebang {
		combined.WriteString(constants.COMBINE_SHEBANG + "\n")
	}

	// Add project header
	combined.WriteString("#\n")
	combined.WriteString("# " + config.Project)

	if config.Version != "" {
		combined.WriteString(" V-" + config.Version)
	}

	combined.WriteString("\n")
	combined.WriteString("#\n\n")

	// Add licence
	if config.Licence != "" {
		if licenceContent, err := os.ReadFile(config.Licence); err == nil {
			for _, line := range strings.Split(string(licenceContent), "\n") {
				if strings.TrimSpace(line) != "" {
					combined.WriteString("# " + line + "\n")
				} else {
					combined.WriteString("#\n")
				}
			}

			combined.WriteString("\n")
		}
	}

	// Process each file
	for _, filename := range files {
		content, err := os.ReadFile(filename)

		if err != nil {
			return fmt.Errorf("reading file '%s': %w", filename, err)
		}

		processedContent := processFileContent(string(content), config)

		if processedContent != "" {
			combined.WriteString(processedContent)

			if !config.StripWhitespace {
				combined.WriteString("\n")
			}
		}
	}

	// Write output file
	err := os.WriteFile(config.Output, []byte(combined.String()), constants.FILE_PERM_EXECUTABLE)

	if err != nil {
		return fmt.Errorf("writing output file '%s': %w", config.Output, err)
	}

	return nil
}

func stripComments(line string) string {
	inString := false
	escaped := false

	for i, char := range line {
		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if char == '#' && !inString {
			return strings.TrimRightFunc(line[:i], func(r rune) bool {
				return r == ' ' || r == '\t'
			})
		}
	}

	return line
}

func processFileContent(content string, config CombineConfig) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var builder strings.Builder

	// Build external lookup map
	externalMap := make(map[string]bool)

	for _, ext := range config.External {
		externalMap[ext] = true
	}

	for scanner.Scan() {
		line := scanner.Text()

		// Skip shebang lines
		if strings.HasPrefix(line, "#!") {
			continue
		}

		// Keep external, remove bundled
		if loadCallRegex.MatchString(line) {
			matches := loadRegex.FindStringSubmatch(line)

			if len(matches) > 1 {
				depPath := matches[1]

				// Keep load() for external dependencies
				if !externalMap[depPath] {
					continue
				}
			} else {
				continue
			}
		}

		// Check if line is comment only before stripping
		trimmedLine := strings.TrimSpace(line)
		isCommentOnly := trimmedLine != "" && strings.HasPrefix(trimmedLine, "#")

		// Strip comments if requested
		if config.StripComments {
			line = stripComments(line)
		}

		// Strip whitespace if requested
		if config.StripWhitespace {
			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			builder.WriteString(line + " ")
		} else {
			if config.StripComments && isCommentOnly {
				continue
			}

			builder.WriteString(line + "\n")
		}
	}

	return builder.String()
}
