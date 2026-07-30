/*
 *
 * The Chippy Disassembler Tool
 *
 */

package main

import (
	"chip-go/internal/bytecode/disassembler"
	"chip-go/internal/compiler"
	"fmt"
	"os"
)

func handleDisasmCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: no file specified")

		os.Exit(1)
	}

	filename := args[0]

	data, err := os.ReadFile(filename)

	if err != nil {
		fmt.Printf("Error: could not read '%s': %v\n", filename, err)

		os.Exit(1)
	}

	chunk, err := compiler.DecodeOrCompile(filename, data)

	if err != nil {
		fmt.Println(err.Error())

		os.Exit(1)
	}

	fmt.Print(disassembler.DisassembleFull(chunk, filename))
}
