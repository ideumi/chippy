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
	inlineCodeRegex = regexp.MustCompile(regexp.QuoteMeta(constants.INLINE_CODE_MARKER) + `([^']+)` + regexp.QuoteMeta(constants.INLINE_CODE_MARKER))
	boldTextRegex   = regexp.MustCompile(regexp.QuoteMeta(constants.BOLD_MARKER) + `([^*]+)` + regexp.QuoteMeta(constants.BOLD_MARKER))
	variableRegex   = regexp.MustCompile(`\$CHIP(VR|CN)`)
	symbolDeclRegex = regexp.MustCompile(regexp.QuoteMeta(constants.DOC_COMMENT_PREFIX) + `\s*` + regexp.QuoteMeta(constants.SYMBOL_DECL_OPEN) + `(.+)` + regexp.QuoteMeta(constants.SYMBOL_DECL_CLOSE))
)

func renderChpDoc(content string) string {
	lines := strings.Split(content, "\n")

	var result strings.Builder

	// Handle code block line numbers with more lineNumWidth
	// needed than 1

	// Are there any code blocks?
	hasCodeBlocks := false
	for _, line := range lines {
		if strings.HasPrefix(line, constants.CODE_BLOCK_DELIMITER) {
			hasCodeBlocks = true
			break
		}
	}

	inCodeBlock := false
	codeLineNum := constants.CODE_LINE_START
	lineNumWidth := 1

	// Only do this if we have code blocks
	if hasCodeBlocks {
		maxLineNum := 0
		tempInCodeBlock := false
		tempLineNum := constants.CODE_LINE_START

		for _, line := range lines {
			if strings.HasPrefix(line, constants.CODE_BLOCK_DELIMITER) {
				tempInCodeBlock = !tempInCodeBlock

				if tempInCodeBlock {
					tempLineNum = constants.CODE_LINE_START
				}

				continue
			}
			if tempInCodeBlock {
				if tempLineNum > maxLineNum {
					maxLineNum = tempLineNum
				}

				tempLineNum++
			}
		}

		// Calculate width needed for line numbers
		lineNumWidth = len(fmt.Sprintf("%d", maxLineNum))
	}

	for _, line := range lines {
		// Code blocks
		if strings.HasPrefix(line, constants.CODE_BLOCK_DELIMITER) {
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
			// Pad
			result.WriteString(fmt.Sprintf("%*d│    %s\n", lineNumWidth, codeLineNum, line))
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

	// If doc/ exists in cwd add it
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

func displayChpDocFile(path string) error {
	content, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	rendered := renderChpDoc(string(content))
	fmt.Print(rendered)

	return nil
}

func showHelp(document, symbol string) {
	// Document is a .chpdoc
	if strings.HasSuffix(document, constants.DOC_FILE_EXTENSION) {
		if err := displayChpDocFile(document); err == nil {
			return
		}
	}

	// Document is a source file
	if strings.HasSuffix(document, constants.FILE_EXT_PROG) ||
		strings.HasSuffix(document, constants.FILE_EXT_HEADER) || strings.Contains(document, "/") {

		if _, err := os.Stat(document); err == nil {
			showFileDocumentation(document, symbol)

			return
		}
	}

	// Document is either in installation or in userdirs ...

	docPaths := getDocPaths()

	if document == "" {
		// Index file - try each path until found
		for _, docPath := range docPaths {
			indexPath := filepath.Join(docPath, constants.DOC_INDEX_FILE)

			if err := displayChpDocFile(indexPath); err == nil {
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

		if err := displayChpDocFile(documentPath); err == nil {
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
			continue
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

type SymbolDoc struct {
	Signature string
	DocBlock  string
}

func parseSymbolDeclaration(line string) (signature string, ok bool) {
	matches := symbolDeclRegex.FindStringSubmatch(line)

	if matches == nil {
		return "", false
	}

	signature = strings.TrimSpace(matches[1])

	return signature, true
}

func extractDocumentation(filepath string) ([]SymbolDoc, error) {
	content, err := os.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var symbols []SymbolDoc

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		trimmed := strings.TrimSpace(line)

		if !strings.HasPrefix(trimmed, constants.DOC_COMMENT_PREFIX) {
			continue
		}

		sig, ok := parseSymbolDeclaration(trimmed)

		if !ok {
			continue
		}

		i++

		var docLines []string

		for i < len(lines) {
			nextLine := lines[i]
			nextTrimmed := strings.TrimSpace(nextLine)

			if !strings.HasPrefix(nextTrimmed, constants.DOC_COMMENT_PREFIX) {
				break
			}

			lineContent := strings.TrimPrefix(nextTrimmed, constants.DOC_COMMENT_PREFIX)

			// Trim spaces
			lineContent = strings.TrimLeft(lineContent, " ")
			docLines = append(docLines, lineContent)
			i++
		}

		symbols = append(symbols, SymbolDoc{
			Signature: sig,
			DocBlock:  strings.Join(docLines, "\n"),
		})
	}

	return symbols, nil
}

func listFileSymbols(symbols []SymbolDoc) {
	fmt.Println("Available documentation")
	fmt.Println()

	if len(symbols) == 0 {
		fmt.Println("(none)")
		fmt.Println()

		return
	}

	for _, sym := range symbols {
		fmt.Println(sym.Signature)
	}

	fmt.Println()
	fmt.Println("Use 'chippy doc <file> <symbol>' for detailed help.")
}

func renderSymbolDoc(sym SymbolDoc) {
	// Render signature
	fmt.Printf("\033[4m%s\033[0m\n\n", sym.Signature)

	rendered := renderChpDoc(sym.DocBlock)
	fmt.Print(rendered)
}

func showFileDocumentation(filepath, symbolSignature string) {
	symbols, err := extractDocumentation(filepath)

	if err != nil {
		fmt.Printf("Error: Cannot read file %s: %s\n", filepath, err.Error())

		os.Exit(1)
	}

	if symbolSignature == "" {
		listFileSymbols(symbols)

		return
	}

	if symbolSignature == constants.SYMBOL_FILTER_ALL {
		if len(symbols) == 0 {
			fmt.Printf("Error: No documentation found in %s\n", filepath)

			os.Exit(1)
		}

		for i, sym := range symbols {
			if i > 0 {
				fmt.Println()
				fmt.Println()
			}

			renderSymbolDoc(sym)
		}
		return
	}

	for _, sym := range symbols {
		if sym.Signature == symbolSignature {
			renderSymbolDoc(sym)

			return
		}
	}

	fmt.Printf("Error: No documentation found for symbol '%s' in %s\n", symbolSignature, filepath)
	os.Exit(1)
}
