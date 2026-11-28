/*
 *
 * Combine Dependency Graph
 *
 */

package combine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// FIXME: This file should probably be optimized, three file reads is not optimal, but oh well.

var (
	loadRegex     = regexp.MustCompile(`load\s*\(\s*"((?:[^"\\]|\\.)*)"\s*\)`)
	loadCallRegex = regexp.MustCompile(`^\s*load\s*\(`)
)

func isBundledLoadLine(line string, external map[string]bool) bool {
	if !loadCallRegex.MatchString(line) {
		return false
	}

	matches := loadRegex.FindStringSubmatch(line)

	if len(matches) <= 1 {
		return true
	}

	return !external[matches[1]]
}

// removeLoadLines removes every bundled load(filename) call.
func removeLoadLines(content string, external map[string]bool) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if isBundledLoadLine(line, external) {
			// Bit of a HACK:
			// Keep line count, so runtime positions map 1:1 to the file on disk
			lines[i] = ""
		}
	}

	return strings.Join(lines, "\n")
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
