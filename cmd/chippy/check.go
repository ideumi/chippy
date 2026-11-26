/*
 *
 * The Chippy Check Tool
 *
 */

package main

import (
	"chip-go/internal/lexer"
	"chip-go/internal/parser"
	"fmt"
	"os"
)

func handleCheckCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: no files specified")

		os.Exit(1)
	}

	if args[0] == "all" {
		dir := "."

		if len(args) > 1 {
			dir = args[1]
		}

		checkAllFiles(dir)
		return
	}

	hasErrors := false

	for _, filename := range args {
		if err := checkFile(filename); err != nil {
			fmt.Printf("%s: %v\n", filename, err)
			hasErrors = true
		} else {
			fmt.Printf("%s: OK\n", filename)
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func checkAllFiles(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Error: directory does not exist: %s\n", dir)

		os.Exit(1)
	}

	files, err := walkChippyFiles(dir)

	if err != nil {
		fmt.Printf("Error: can't walk directory: %v\n", err)

		os.Exit(1)
	}

	hasErrors := false

	for _, filename := range files {
		if err := checkFile(filename); err != nil {
			fmt.Printf("%s: %v\n", filename, err)

			hasErrors = true
		} else {
			fmt.Printf("%s: OK\n", filename)
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

func checkFile(filename string) error {
	content, err := os.ReadFile(filename)

	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	lex := lexer.NewLexer(filename, string(content))
	tokens, err := lex.MakeTokens()

	if err != nil {
		return err
	}

	parseResult := parser.NewParser(tokens).Parse()

	if parseResult.GetError() != nil {
		return parseResult.GetError()
	}

	return nil
}
