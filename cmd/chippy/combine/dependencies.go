/*
 *
 * The Chippy Combine Tool - Dependency Resolution
 *
 */

package combine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var (
	loadRegex      = regexp.MustCompile(`\bload\s*\(\s*"((?:[^"\\]|\\.)*)"\s*\)`)
	loadCallRegex  = regexp.MustCompile(`^\s*load\s*\(`)
	ploadRegex     = regexp.MustCompile(`\bpload\s*\(\s*"((?:[^"\\]|\\.)*)"\s*\)`)
	ploadCallRegex = regexp.MustCompile(`^\s*pload\s*\(`)
)

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

func collectPlugins(programFiles []string, pluginPaths []string, externalPlugins []string) ([]string, []string, error) {
	seenPlugins := make(map[string]bool)
	skippedPlugins := make(map[string]bool)

	var plugins []string

	// Build external plugins lookup map
	externalPluginMap := make(map[string]bool)

	for _, ext := range externalPlugins {
		externalPluginMap[ext] = true
	}

	// Scan all program files for pload()
	for _, file := range programFiles {
		content, err := os.ReadFile(file)

		if err != nil {
			return nil, nil, fmt.Errorf("reading file '%s': %w", file, err)
		}

		matches := ploadRegex.FindAllStringSubmatch(string(content), -1)

		for _, match := range matches {
			if len(match) > 1 {
				pluginName := match[1]

				// Check if plugin is external
				if externalPluginMap[pluginName] {
					skippedPlugins[pluginName] = true

					continue
				}

				// Resolve plugin path
				resolvedPath := resolvePath(pluginName, pluginPaths)

				if resolvedPath == "" {
					return nil, nil, fmt.Errorf("plugin '%s' not found in plugin paths %v", pluginName, pluginPaths)
				}

				if !seenPlugins[resolvedPath] {
					seenPlugins[resolvedPath] = true
					plugins = append(plugins, resolvedPath)
				}
			}
		}
	}

	// Convert skipped map to slice
	var skipped []string

	for name := range skippedPlugins {
		skipped = append(skipped, name)
	}

	sort.Strings(skipped)

	return plugins, skipped, nil
}

func collectAssets(assetPaths []string) ([]string, error) {
	var assets []string

	for _, assetPath := range assetPaths {
		if _, err := os.Stat(assetPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("asset '%s' not found", assetPath)
		}

		assets = append(assets, assetPath)
	}

	return assets, nil
}
