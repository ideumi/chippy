/*
*
* The Chippy Documentation Tool
*
 */

package main

import (
	"chip-go/internal/constants"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Much fast
var (
	inlineCodeRegex = regexp.MustCompile(`''([^']+)''`)
	boldTextRegex   = regexp.MustCompile(`\*([^*]+)\*`)
	variableRegex   = regexp.MustCompile(`\$CHIP(VR|CN)`)
)

func renderChpDoc(content string) string {
	lines := strings.Split(content, "\n")

	var result strings.Builder

	inCodeBlock := false
	codeLineNum := constants.CODE_LINE_START

	for _, line := range lines {
		// Code blocks
		if strings.HasPrefix(line, "'''") {
			inCodeBlock = !inCodeBlock

			if inCodeBlock {
				result.WriteString("\n")
				codeLineNum = constants.CODE_LINE_START
			} else {
				result.WriteString("\n")
			}

			continue
		}

		// Code block line numbers
		if inCodeBlock {
			result.WriteString(fmt.Sprintf("%d│    %s\n", codeLineNum, line))
			codeLineNum++

			continue
		}

		processedLine := processInlineFormatting(line)

		if strings.TrimSpace(line) == "" {
			result.WriteString("\n")
		} else {
			result.WriteString(processedLine + "\n")
		}
	}

	return result.String()
}

func processInlineFormatting(line string) string {
	// Variable substitution
	line = variableRegex.ReplaceAllStringFunc(line, func(match string) string {
		switch match {
		case "$CHIPVR":
			return constants.STR_LPLVR
		case "$CHIPCN":
			return constants.STR_LPLCN
		default:
			return match
		}
	})

	// Cyan text for code
	line = inlineCodeRegex.ReplaceAllString(line, "\033[36m$1\033[0m")

	// Bold text
	line = boldTextRegex.ReplaceAllString(line, "\033[1m$1\033[0m")

	return line
}

func getDocPaths() []string {
	var paths []string

	// If dev/ exists in cwd add it
	if _, err := os.Stat(constants.DOC_DIR_LOCAL); err == nil {
		paths = append(paths, constants.DOC_DIR_LOCAL)
	}

	// Check user local directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		userDocPath := filepath.Join(homeDir, constants.DOC_DIR_USER)

		if _, err := os.Stat(userDocPath); err == nil {
			paths = append(paths, userDocPath)
		}
	}

	// Add production path based on OS
	if runtime.GOOS == "android" {
		paths = append(paths, constants.DOC_DIR_PRODUCTION_TERMUX)
	} else {
		paths = append(paths, constants.DOC_DIR_PRODUCTION)
	}

	return paths
}

func displayMarkdownFile(path string) error {
	content, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	rendered := renderChpDoc(string(content))
	fmt.Print(rendered)

	return nil
}

func showHelp(document string) {
	docPaths := getDocPaths()

	if document == "" {
		// Index file - try each path until found
		for _, docPath := range docPaths {
			indexPath := filepath.Join(docPath, constants.DOC_INDEX_FILE)

			if err := displayMarkdownFile(indexPath); err == nil {
				return
			}
		}

		fmt.Printf("Error: Index document not found in any documentation directory\n")

		os.Exit(1)
	}

	if document == "list" {
		listFunctions(docPaths)
		return
	}

	// Try each path until document is found
	for _, docPath := range docPaths {
		documentPath := filepath.Join(docPath, document+constants.DOC_FILE_EXTENSION)

		if err := displayMarkdownFile(documentPath); err == nil {
			return
		}
	}

	fmt.Printf("Error: No document available for '%s'\n", document)
	os.Exit(1)
}

func listFunctions(docPaths []string) {
	fmt.Println("Available documentation")
	fmt.Println()

	topicsSet := make(map[string]bool)
	var topics []string

	// Collect all unique topics from all documentation directories
	for _, docPath := range docPaths {
		files, err := os.ReadDir(docPath)

		if err != nil {
			continue // Skip directories that can't be read
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), constants.DOC_FILE_EXTENSION) && file.Name() != constants.DOC_INDEX_FILE {
				name := strings.TrimSuffix(file.Name(), constants.DOC_FILE_EXTENSION)

				if !topicsSet[name] {
					topicsSet[name] = true
					topics = append(topics, name)
				}
			}
		}
	}

	const colWidth = constants.COLUMN_WIDTH
	const numCols = constants.NUM_COLUMNS

	for i, topic := range topics {
		fmt.Printf("%-*s", colWidth, topic)

		if (i+1)%numCols == 0 {
			fmt.Println()
		}
	}

	if len(topics)%numCols != 0 {
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("Use 'chippy doc <document>' for detailed help.")
}
