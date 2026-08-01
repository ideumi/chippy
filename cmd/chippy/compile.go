/*
 *
 * The Chippy Compile Tool
 *
 */

package main

import (
	"chip-go/internal/compiler"
	"chip-go/internal/constants"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func handleCompileCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: no source file specified")

		os.Exit(1)
	}

	entry := args[0]

	data, err := os.ReadFile(entry)

	if err != nil {
		fmt.Printf("Error: could not read '%s': %v\n", entry, err)

		os.Exit(1)
	}

	base := filepath.Base(entry)
	output := strings.TrimSuffix(base, filepath.Ext(base))

	if len(args) > 1 {
		output = args[1]
	}

	compiled, err := compiler.CompileToBytecode(entry, string(data), true)

	if err != nil {
		fmt.Println(err.Error())

		os.Exit(1)
	}

	if err := os.WriteFile(output, compiled, constants.FILE_PERM_EXECUTABLE); err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}

	fmt.Printf("Successfully created %s\n", output)
}
