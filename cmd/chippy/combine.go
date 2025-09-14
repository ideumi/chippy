/*
 *
 * The Chippy Combine Tool
 *
 */

package main

import (
	"bufio"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/roadrunner"
	"chip-go/internal/values"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	loadRegex     = regexp.MustCompile(`load\s*\(\s*"((?:[^"\\]|\\.)*)"\s*\)`)
	loadCallRegex = regexp.MustCompile(`^\s*load\s*\(`)
)

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

if homePath != err
{
    Paths = Paths + [homePath + "/.local/share/chiplang/lib"];
}

Paths = Paths + ["/usr/lib/chiplang"];

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

	// Build dependency graph from entry point
	deps, err := buildDependencyGraph(config.Source, config.Paths)

	if err != nil {
		return err
	}

	// Show bundling information
	fmt.Printf("Bundling %s V-%s:\n", config.Project, config.Version)
	fmt.Printf("Output: %s\n", config.Output)
	fmt.Printf("\nFiles bundled:\n")

	for i, file := range deps {
		fmt.Printf("  %d. %s\n", i+1, file)
	}

	fmt.Printf("\nTotal: %d source(s)\n", len(deps))
	fmt.Println()

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
	StripComments   bool
	StripWhitespace bool
	AddShebang      bool
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

	// Flags
	if val := ctx.SymbolTable.Get(constants.CONFIG_STRIP_COMMENTS); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.StripComments = num.Value != 0
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_STRIP_WHITESPACE); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.StripWhitespace = num.Value != 0
		}
	}

	if val := ctx.SymbolTable.Get(constants.CONFIG_ADD_SHEBANG); val != nil {
		if num, ok := val.(*values.Number); ok {
			config.AddShebang = num.Value != 0
		}
	}

	return config
}

func buildDependencyGraph(source string, paths []string) ([]string, error) {
	seen := make(map[string]bool)
	var result []string

	// Process dependencies starting from entry point
	if err := processDependencies(source, paths, seen, &result); err != nil {
		return nil, fmt.Errorf("processing dependencies: %w", err)
	}

	return result, nil
}

func processDependencies(filename string, paths []string, seen map[string]bool, result *[]string) error {
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

			// Resolve dependency through search paths
			resolvedPath := resolvePath(depPath, paths)

			if resolvedPath == "" {
				return fmt.Errorf("dependency '%s' not found in search paths %v", depPath, paths)
			}

			if !seen[resolvedPath] {
				if err := processDependencies(resolvedPath, paths, seen, result); err != nil {
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

	for scanner.Scan() {
		line := scanner.Text()

		// Skip shebang lines
		if strings.HasPrefix(line, "#!") {
			continue
		}

		// Remove load() calls
		if loadCallRegex.MatchString(line) {
			continue
		}

		// Strip comments if requested but not inside string literals
		if config.StripComments {
			line = stripComments(line)
		}

		// Strip whitespace if requested
		if config.StripWhitespace {
			line = strings.TrimSpace(line)

			if line == "" {
				continue
			}

			builder.WriteString(line)
		} else {
			if line != "" || !config.StripWhitespace {
				builder.WriteString(line + "\n")
			}
		}
	}

	return builder.String()
}
