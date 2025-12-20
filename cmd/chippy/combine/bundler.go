/*
 *
 * The Chippy Combine Tool - Bundler
 *
 */

package combine

import (
	"bufio"
	"chip-go/internal/bundle"
	"chip-go/internal/constants"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func generateCombinedFile(config CombineConfig, files []string, plugins []string, assets []string) error {
	if config.ProduceBundle {
		return generateBinaryBundle(config, files, plugins, assets)
	} else {
		return generateTextBundle(config, files)
	}
}

func generateBinaryBundle(config CombineConfig, files []string, plugins []string, assets []string) error {
	bundleID, err := generateBundleID()

	if err != nil {
		return fmt.Errorf("generating bundle ID: %w", err)
	}

	program, err := generateBundledProgram(config, files, bundleID)

	if err != nil {
		return fmt.Errorf("generating bundled program: %w", err)
	}

	// Build bundle
	b := &bundle.Bundle{
		Project:         config.Project,
		Version:         config.Version,
		ChipLangVersion: constants.STR_LPLVR,
		BundleID:        bundleID,
		Program:         program,
	}

	// Check for duplicate file basenames
	fileBasenames := make(map[string]string)

	for _, pluginPath := range plugins {
		basename := filepath.Base(pluginPath)

		if existingPath, exists := fileBasenames[basename]; exists {
			return fmt.Errorf("duplicate file basename '%s' from '%s' and '%s'", basename, existingPath, pluginPath)
		}

		fileBasenames[basename] = pluginPath
	}

	for _, assetPath := range assets {
		basename := filepath.Base(assetPath)

		if existingPath, exists := fileBasenames[basename]; exists {
			return fmt.Errorf("duplicate file basename '%s' from '%s' and '%s'", basename, existingPath, assetPath)
		}

		fileBasenames[basename] = assetPath
	}

	// Add plugin files
	for _, pluginPath := range plugins {
		data, err := os.ReadFile(pluginPath)

		if err != nil {
			return fmt.Errorf("reading plugin '%s': %w", pluginPath, err)
		}

		b.Files = append(b.Files, bundle.BundleFile{
			Name: filepath.Base(pluginPath),
			Type: bundle.FileTypePlugin,
			Data: data,
		})
	}

	// Add asset files
	for _, assetPath := range assets {
		data, err := os.ReadFile(assetPath)

		if err != nil {
			return fmt.Errorf("reading asset '%s': %w", assetPath, err)
		}

		b.Files = append(b.Files, bundle.BundleFile{
			Name: filepath.Base(assetPath),
			Type: bundle.FileTypeAsset,
			Data: data,
		})
	}

	// Auto detect if bundle has plugins
	b.DetectPlugins()

	if err := bundle.WriteBundle(config.Output, b); err != nil {
		return fmt.Errorf("writing bundle: %w", err)
	}

	return nil
}

func generateTextBundle(config CombineConfig, files []string) error {
	var combined strings.Builder

	// Add shebang
	combined.WriteString(constants.COMBINE_SHEBANG + "\n")

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

	for _, filename := range files {
		content, err := os.ReadFile(filename)

		if err != nil {
			return fmt.Errorf("reading file '%s': %w", filename, err)
		}

		processedContent := processFileContent(string(content), config, nil)

		if processedContent != "" {
			combined.WriteString(processedContent)

			if !config.StripWhitespace {
				combined.WriteString("\n")
			}
		}
	}

	err := os.WriteFile(config.Output, []byte(combined.String()), constants.FILE_PERM_EXECUTABLE)

	if err != nil {
		return fmt.Errorf("writing output file '%s': %w", config.Output, err)
	}

	return nil
}

func generateBundledProgram(config CombineConfig, files []string, bundleID uint64) (string, error) {
	var combined strings.Builder

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

	for _, filename := range files {
		content, err := os.ReadFile(filename)

		if err != nil {
			return "", fmt.Errorf("reading file '%s': %w", filename, err)
		}

		processedContent := processFileContent(string(content), config, &bundleID)

		if processedContent != "" {
			combined.WriteString(processedContent)

			if !config.StripWhitespace {
				combined.WriteString("\n")
			}
		}
	}

	return combined.String(), nil
}

// bundleID is nil for text bundles, non-nil for binary bundles
func processFileContent(content string, config CombineConfig, bundleID *uint64) string {
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

		// Keep external load(), remove bundled load()
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

		// Rewrite pload() calls with bundle extraction path
		if bundleID != nil && ploadCallRegex.MatchString(line) {
			matches := ploadRegex.FindStringSubmatch(line)

			if len(matches) > 1 {
				originalPath := matches[1]
				pluginName := filepath.Base(originalPath)

				// Rewrite to bundle extraction placeholder
				placeholder := fmt.Sprintf("$CHIPBIN_EXTRACT/%d/%s", *bundleID, pluginName)

				line = strings.Replace(line,
					fmt.Sprintf(`pload("%s")`, originalPath),
					fmt.Sprintf(`pload("%s")`, placeholder),
					1)
			}
		}

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

			builder.WriteString(line)
		} else {
			if line != "" || !config.StripWhitespace {
				builder.WriteString(line + "\n")
			}
		}
	}

	return builder.String()
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

func generateBundleID() (uint64, error) {
	timestamp := uint64(time.Now().UnixNano())

	var randomBytes [8]byte

	if _, err := rand.Read(randomBytes[:]); err != nil {
		return 0, fmt.Errorf("generating random data: %w", err)
	}

	randomPart := binary.LittleEndian.Uint64(randomBytes[:])

	id := timestamp ^ randomPart

	return id, nil
}
