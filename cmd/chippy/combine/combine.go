/*
 *
 * Combine Bundler
 *
 */

package combine

import (
	"bufio"
	"chip-go/cmd/chippy/formatter"
	"chip-go/internal/compiler"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	errBundleErrors     = errors.New(constants.COMBINE_E_BUNDLE_ERRORS)
	errBundleCollisions = errors.New(constants.COMBINE_E_BUNDLE_COLLISIONS)
)

// GenerateTemplate writes a fresh combine configuration file to the current
// directory. It refuses to overwrite an existing one.
func GenerateTemplate() error {
	// Check if combine.chp already exists
	if _, err := os.Stat(constants.COMBINE_DEFAULT_FILENAME); err == nil {
		return fmt.Errorf("file '%s' already exists, rejected", constants.COMBINE_DEFAULT_FILENAME)
	}

	template := constants.CHIPPY_SHEBANG + `
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

if not iserr(homePath)
{
	Paths = Paths + [homePath + "/.local/share/chippy/lib"];
}

Paths = Paths + ["/usr/lib/chippy"];

# Dependencies combine shouldn't bundle and just keep as loads

var External = [];

# Options

var Compile = true;
var AddShebang = true;

var StripComments = true;
var StripWhitespace = true;
`

	err := os.WriteFile(constants.COMBINE_DEFAULT_FILENAME, []byte(template), constants.FILE_PERM_READABLE)

	if err != nil {
		return fmt.Errorf("creating template file: %w", err)
	}

	fmt.Printf("Created %s\n", constants.COMBINE_DEFAULT_FILENAME)

	return nil
}

// Execute bundles the config's program and its dependencies into a single self-
// contained executable.
func Execute(combineFile string) error {
	config, err := loadCombineConfig(combineFile)

	if err != nil {
		return err
	}

	if config.Project == "" {
		return fmt.Errorf("project name not specified in '%s'", combineFile)
	}

	if config.Output == "" {
		config.Output = config.Project
	}

	if err := validateCombineSource(config, combineFile); err != nil {
		return err
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
	validation, err := ValidateFiles(deps)

	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Report parse errors
	if len(validation.Errors) > 0 {
		reportParseErrors(validation.Errors)

		return errBundleErrors
	}

	// Report symbol collisions
	if len(validation.Collisions) > 0 {
		reportCollisions(validation.Collisions)

		return errBundleCollisions
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

func generateCombinedFile(config CombineConfig, files []string) error {
	var combined strings.Builder

	// Add shebang if requested
	if config.AddShebang {
		combined.WriteString(constants.CHIPPY_SHEBANG + "\n")
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

	externalMap := make(map[string]bool)

	for _, ext := range config.External {
		externalMap[ext] = true
	}

	// Process each file
	for _, filename := range files {
		content, err := os.ReadFile(filename)

		if err != nil {
			return fmt.Errorf("reading file '%s': %w", filename, err)
		}

		processedContent := processFileContent(removeLoadLines(string(content), externalMap), config)

		if processedContent != "" {
			combined.WriteString(processedContent)

			if !config.StripWhitespace {
				combined.WriteString("\n")
			}
		}
	}

	// Normalize the tail to a single newline.
	output := strings.TrimRight(combined.String(), " \t\r\n") + "\n"

	// Run the formatter on readable output
	if !config.StripWhitespace {
		if formatted, err := formatter.Format(config.Output, output); err == nil {
			output = formatted
		}
	}

	if config.Compile {
		return writeCompiledBundle(config.Output, output, config.AddShebang)
	}

	if err := os.WriteFile(config.Output, []byte(output), constants.FILE_PERM_EXECUTABLE); err != nil {
		return fmt.Errorf("writing output file '%s': %w", config.Output, err)
	}

	return nil
}

func writeCompiledBundle(outputPath, source string, addShebang bool) error {
	compiled, err := compiler.CompileToBytecode(outputPath, source, addShebang)

	if err != nil {
		return fmt.Errorf("compiling bundle: %w", err)
	}

	if err := os.WriteFile(outputPath, compiled, constants.FILE_PERM_EXECUTABLE); err != nil {
		return fmt.Errorf("writing output file '%s': %w", outputPath, err)
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
			return strings.TrimRightFunc(line[:i], func(trailingChar rune) bool {
				return trailingChar == ' ' || trailingChar == '\t'
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

// exitWithResult ends the process with the value's integer as the exit code.
func exitWithResult(value values.Value) {
	code, _ := value.AsInt()

	os.Exit(int(code))
}
