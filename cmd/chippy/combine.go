/*
 *
 * The Chippy Combine Tool
 *
 */

package main

import (
	"chip-go/cmd/chippy/combine"
	"chip-go/internal/constants"
	"fmt"
	"os"
)

func handleCombineCommand(args []string) {
	if len(args) > 0 && args[0] == constants.COMBINE_NEW_COMMAND {
		if err := combine.GenerateTemplate(); err != nil {
			fmt.Printf("Error: %v\n", err)

			os.Exit(1)
		}

		return
	}

	if len(args) > 0 && args[0] == constants.COMBINE_TEST_COMMAND {
		handleCombineTestCommand("", args[1:])

		return
	}

	if len(args) > 0 && args[0] == constants.COMBINE_RUN_COMMAND {
		handleCombineRunCommand("", args[1:])

		return
	}

	if len(args) > 1 && args[1] == constants.COMBINE_TEST_COMMAND {
		handleCombineTestCommand(args[0], args[2:])

		return
	}

	if len(args) > 1 && args[1] == constants.COMBINE_RUN_COMMAND {
		handleCombineRunCommand(args[0], args[2:])

		return
	}

	combineFile := constants.COMBINE_DEFAULT_FILENAME

	if len(args) > 0 {
		combineFile = args[0]
	}

	if _, err := os.Stat(combineFile); os.IsNotExist(err) {
		fmt.Printf("Error: combine file '%s' not found\n", combineFile)

		os.Exit(1)
	}

	if err := combine.Execute(combineFile); err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
}

func handleCombineTestCommand(cfgFile string, rest []string) {
	if cfgFile == "" {
		cfgFile = constants.COMBINE_DEFAULT_FILENAME
	}

	if err := combine.Test(cfgFile, rest); err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
}

func handleCombineRunCommand(cfgFile string, rest []string) {
	if cfgFile == "" {
		cfgFile = constants.COMBINE_DEFAULT_FILENAME
	}

	if err := combine.Run(cfgFile, rest); err != nil {
		fmt.Printf("Error: %v\n", err)

		os.Exit(1)
	}
}
