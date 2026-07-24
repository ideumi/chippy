/*
 *
 * The Chippy REPL
 *
 */

package main

import (
	"chip-go/internal/builtins"
	"chip-go/internal/constants"
	"chip-go/internal/modena"
	"chip-go/thirdparty/readline"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if runtime.GOOS != "linux" && runtime.GOOS != "android" {
		fmt.Println("Unsupported operating system")
		os.Exit(1)
	}

	args := os.Args[1:]

	// Handle help flag
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		showHelp("chippy", "")
		return
	}

	// Handle doc command
	if len(args) > 0 && args[0] == "doc" {
		if len(args) == 1 {
			showHelp("", "")
		} else if len(args) == 2 {
			showHelp(args[1], "")
		} else {
			showHelp(args[1], args[2])
		}
		return
	}

	// Handle combine command
	if len(args) > 0 && args[0] == "combine" {
		handleCombineCommand(args[1:])
		return
	}

	// Handle check command
	if len(args) > 0 && args[0] == "check" {
		handleCheckCommand(args[1:])
		return
	}

	// Handle format command
	if len(args) > 0 && args[0] == "format" {
		handleFormatCommand(args[1:])
		return
	}

	// Handle compile command
	if len(args) > 0 && args[0] == "compile" {
		handleCompileCommand(args[1:])
		return
	}

	// Handle disasm command
	if len(args) > 0 && args[0] == "disasm" {
		handleDisasmCommand(args[1:])
		return
	}

	mod := modena.New()

	// Start REPL
	if len(args) == 0 {
		builtins.SetGlobalArgs([]string{})

		runREPL(mod)
	} else if len(args) >= 1 && (args[0] == "-v" || args[0] == "--version") {
		fmt.Printf("%s %s %s\n",
			constants.STR_LPLVR, constants.VERSION_DATE, constants.STR_LPLCN)
	} else if len(args) >= 2 && (args[0] == "-r" || args[0] == "--run") {
		builtins.SetGlobalArgs([]string{}) // No args for -r / --run mode

		runCommand(mod, args[1])
	} else {
		scriptArgs := []string{}

		if len(args) > 1 {
			scriptArgs = args[1:]
		}

		builtins.SetGlobalArgs(scriptArgs)
		runScript(mod, args[0])
	}
}

func runREPL(mod *modena.Modena) {
	fmt.Printf("chippy V-%s '%s' from %s on %s-%s.\n",
		constants.STR_LPLVR, constants.STR_LPLCN, constants.VERSION_DATE, runtime.GOOS, runtime.GOARCH)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	historyFile := filepath.Join(homeDir, constants.HIST_FILE)

	cwd, _ := os.Getwd()

	prompt := fmt.Sprintf("\033[93m%s\033[0m %s \033[93m➜\033[0m ",
		constants.CLI_CONTEXT_DISPLAY_NAME_FN, cwd)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     historyFile,
		HistoryLimit:    1000,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		fmt.Printf("Error setting up REPL: %v\n", err)
		return
	}

	defer rl.Close()

	for {
		cwd, _ := os.Getwd()

		newPrompt := fmt.Sprintf("\033[93m%s\033[0m %s \033[93m➜\033[0m ",
			constants.CLI_CONTEXT_DISPLAY_NAME_FN, cwd)

		rl.SetPrompt(newPrompt)

		text, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				fmt.Println("\nKeyboard interrupt")
				os.Exit(1)
			} else {
				break
			}
		}

		text = strings.TrimSpace(text)
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		if text == "off" || text == "exit" {
			break
		}

		if text == "history" {
			showHistory(historyFile)
			continue
		}

		result, err := mod.Run(constants.CLI_CONTEXT_DISPLAY_NAME_FN, text)

		if err != nil {
			fmt.Println(err.Error())
		} else if result != nil {
			// Print result immediately and clear reference for GC
			fmt.Println(result.String())

			// Clear the result reference immediately after printing
			result = nil
		}
	}
}

func showHistory(historyFile string) {
	content, err := os.ReadFile(historyFile)
	if err != nil {
		fmt.Println("No history available")
		return
	}

	lines := strings.Split(string(content), "\n")
	start := len(lines) - 21

	if start < 0 {
		start = 0
	}

	fmt.Println("Recent command history:")
	for i := start; i < len(lines)-1; i++ {
		if strings.TrimSpace(lines[i]) != "" {
			fmt.Printf("%3d: %s\n", i-start+1, lines[i])
		}
	}
}

func runCommand(mod *modena.Modena, command string) {
	result, err := mod.Run("<command>", command)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Print result if not null and not empty
	if result != nil {
		fmt.Println(result.String())

		// Clear the result reference immediately after printing
		result = nil
	}
}

func runScript(mod *modena.Modena, filename string) {
	data, err := os.ReadFile(filename)

	if err != nil {
		fmt.Printf("Error: could not read '%s': %v\n", filename, err)
		os.Exit(1)
	}

	if _, err := mod.Run(filename, string(data)); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
