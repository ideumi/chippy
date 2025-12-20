/*
 *
 * The Chippy Bundle Tool
 *
 */

package main

import (
	"chip-go/internal/bundle"
	"chip-go/internal/constants"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handleBundleCommand(args []string) {
	if len(args) == 0 {
		fmt.Printf("Error: no subcommand specified\n")

		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "show":
		if len(args) < 2 {
			fmt.Printf("Error: no bundle file specified\n")

			os.Exit(1)
		}

		handleBundleShow(args[1])
	case "extract":
		if len(args) < 2 {
			fmt.Printf("Error: no bundle file specified\n")

			os.Exit(1)
		}

		handleBundleExtract(args[1])
	default:
		fmt.Printf("Error: unknown subcommand '%s'\n", subcommand)

		os.Exit(1)
	}
}

func handleBundleShow(bundlePath string) {
	b, err := bundle.ReadBundle(bundlePath)

	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	fmt.Printf("Project: %s", b.Project)

	if b.Version != "" {
		fmt.Printf(" V-%s", b.Version)
	}

	fmt.Println()

	fmt.Printf("ChipLang Version: V-%s\n", b.ChipLangVersion)
	fmt.Printf("Bundle ID: %d\n", b.BundleID)

	if b.HasPlugins {
		fmt.Printf("Platform: %s-%s (has plugins)\n", b.TargetOS, b.TargetArch)
	}

	fmt.Printf("\nProgram: %s\n", formatSize(len(b.Program)))

	if len(b.Files) > 0 {
		fmt.Printf("\nFiles (%d):\n", len(b.Files))

		for _, file := range b.Files {
			fmt.Printf("  %-40s %s\n", file.Name, formatSize(len(file.Data)))
		}
	}
}

func handleBundleExtract(bundlePath string) {
	b, err := bundle.ReadBundle(bundlePath)

	if err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	baseName := strings.TrimSuffix(filepath.Base(bundlePath), filepath.Ext(bundlePath))
	extractDir := baseName + "-extracted"

	if _, err := os.Stat(extractDir); err == nil {
		fmt.Printf("Error: directory or file '%s' already exists, rejecting...\n", extractDir)

		os.Exit(1)
	}

	if err := os.Mkdir(extractDir, constants.FILE_PERM_EXECUTABLE); err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	fmt.Printf("Extracting to %s/\n", extractDir)

	fileCount := 0

	programPath := filepath.Join(extractDir, "program.chp")

	if err := os.WriteFile(programPath, []byte(b.Program), constants.FILE_PERM_READABLE); err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	fmt.Printf("  Wrote program.chp (%s)\n", formatSize(len(b.Program)))
	fileCount++

	for _, file := range b.Files {
		if err := bundle.ExtractFile(file, extractDir); err != nil {
			fmt.Printf("Error: %v\n", err)

			os.Exit(1)
		}

		fmt.Printf("  Wrote %s (%s)\n", file.Name, formatSize(len(file.Data)))
		fileCount++
	}

	fmt.Printf("Extracted %d file(s)\n", fileCount)
}

func formatSize(bytes int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < MB {
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	} else if bytes < GB {
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	}

	return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
}
