/*
 *
 * The Chippy Format Tool
 *
 */

package main

import (
	"chip-go/cmd/chippy/formatter"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handleFormatCommand(args []string) {
	checkOnly := false

	if len(args) > 0 && args[0] == "check" {
		checkOnly = true
		args = args[1:]
	}

	files := args

	if len(files) == 0 {
		fmt.Println("Error: no files specified")

		os.Exit(1)
	}

	if files[0] == "all" {
		dir := "."

		if len(files) > 1 {
			dir = files[1]
		}

		formatAllFiles(dir, checkOnly)
		return
	}

	processFiles(files, checkOnly)
}

func formatAllFiles(dir string, checkOnly bool) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Error: directory does not exist: %s\n", dir)

		os.Exit(1)
	}

	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".chp") || strings.HasSuffix(path, ".chh")) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error: can't walk directory: %v\n", err)

		os.Exit(1)
	}

	processFiles(files, checkOnly)
}

func processFiles(files []string, checkOnly bool) {
	hasErrors := false
	needsFormat := false

	for _, filename := range files {
		changed, err := formatFile(filename, checkOnly)

		if err != nil {
			fmt.Printf("%s: %v\n", filename, err)
			hasErrors = true
		} else if changed {
			needsFormat = true
			fmt.Println(filename)
		}
	}

	if hasErrors || (checkOnly && needsFormat) {
		os.Exit(1)
	}
}

// formatFile formats one file and reports whether it changed. Nothing is written
// to disk if checkOnly is set.
func formatFile(filename string, checkOnly bool) (bool, error) {
	content, err := os.ReadFile(filename)

	if err != nil {
		return false, fmt.Errorf("reading file: %w", err)
	}

	formatted, err := formatter.Format(filename, string(content))

	if err != nil {
		return false, err
	}

	if formatted == string(content) {
		return false, nil
	}

	if checkOnly {
		return true, nil
	}

	perm := os.FileMode(0644)

	if info, statErr := os.Stat(filename); statErr == nil {
		perm = info.Mode().Perm()
	}

	if err := os.WriteFile(filename, []byte(formatted), perm); err != nil {
		return false, fmt.Errorf("writing file: %w", err)
	}

	return true, nil
}
