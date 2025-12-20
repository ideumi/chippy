/*
 *
 * The Chippy REPL
 *
 */

package main

import (
	"chip-go/cmd/chippy/combine"
	"chip-go/internal/builtins"
	"chip-go/internal/bundle"
	"chip-go/internal/constants"
	"chip-go/internal/roadrunner"
	"chip-go/internal/values"
	"chip-go/thirdparty/readline"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" &&
		runtime.GOOS != "freebsd" && runtime.GOOS != "openbsd" &&
		runtime.GOOS != "android" {
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
		combine.HandleCombineCommand(args[1:])
		return
	}

	// Handle bundle command
	if len(args) > 0 && args[0] == "bundle" {
		handleBundleCommand(args[1:])
		return
	}

	// Handle check command
	if len(args) > 0 && args[0] == "check" {
		handleCheckCommand(args[1:])
		return
	}

	rr := roadrunner.NewRoadRunner2()

	// Start REPL
	if len(args) == 0 {
		builtins.SetGlobalArgs([]string{})

		runREPL(rr)
	} else if len(args) >= 1 && (args[0] == "-v" || args[0] == "--version") {
		fmt.Printf("%s %s %s\n",
			constants.STR_LPLVR, constants.VERSION_DATE, constants.STR_LPLCN)
	} else if len(args) >= 2 && (args[0] == "-r" || args[0] == "--run") {
		builtins.SetGlobalArgs([]string{}) // No args for -r / --run mode

		runCommand(rr, args[1])
	} else {
		programArgs := []string{}

		if len(args) > 1 {
			programArgs = args[1:]
		}

		builtins.SetGlobalArgs(programArgs)
		runProgram(rr, args[0])
	}
}

func runREPL(rr *roadrunner.RoadRunner2) {
	fmt.Printf("chippy (ChipLang) V-%s '%s' from %s on %s-%s.\n",
		constants.STR_LPLVR, constants.STR_LPLCN, constants.VERSION_DATE, runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS != "linux" && runtime.GOOS != "android" {
		// TODO: Remove this, once testing on the other UNIXes is complete.
		fmt.Printf("WARNING: ChipLang has not been tested on %s, tread lightly.\n", runtime.GOOS)
	}

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

		result, err := rr.Run(constants.CLI_CONTEXT_DISPLAY_NAME_FN, text)

		if err != nil {
			fmt.Println(err.Error())
		} else if result != nil {
			// Print result immediately and clear reference for GC
			if list, ok := result.(*values.List); ok && len(list.Elements) == 1 {
				fmt.Println(list.Elements[0].String())
			} else {
				fmt.Println(result.String())
			}

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

func runCommand(rr *roadrunner.RoadRunner2, command string) {
	result, err := rr.Run("<command>", command)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Print result if not null and not empty
	if result != nil {
		if list, ok := result.(*values.List); ok && len(list.Elements) == 1 {
			fmt.Println(list.Elements[0].String())
		} else {
			fmt.Println(result.String())
		}

		// Clear the result reference immediately after printing
		result = nil
	}
}

func runProgram(rr *roadrunner.RoadRunner2, filename string) {
	if isBundle(filename) {
		if err := runBundle(filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Normal program execution
	chpCode := fmt.Sprintf("load(\"%s\");", filename)

	_, err := rr.Run("<"+filename+">", chpCode)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func isBundle(path string) bool {
	f, err := os.Open(path)

	if err != nil {
		return false
	}

	defer f.Close()

	f.Seek(int64(len(bundle.ShebangLine)), io.SeekStart)

	// Check for CHIPBIN magic bytes
	magic := make([]byte, 7)

	if _, err := f.Read(magic); err != nil {
		return false
	}

	return string(magic) == bundle.MagicBytes
}

func runBundle(bundlePath string) error {
	b, err := bundle.ReadBundle(bundlePath)

	if err != nil {
		return fmt.Errorf("reading bundle: %w", err)
	}

	extractDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("chipbin-%d-", b.BundleID))

	if err != nil {
		return fmt.Errorf("creating extraction directory: %w", err)
	}

	defer os.RemoveAll(extractDir)

	if err := b.Validate(); err != nil {
		return err
	}

	if err := b.Extract(extractDir); err != nil {
		return fmt.Errorf("extracting bundle: %w", err)
	}

	actualProgram := b.GetProgram(extractDir)

	rr := roadrunner.NewRoadRunner2()
	rr.SetBundleConstants(extractDir, b.BundleID)

	_, err = rr.Run(bundlePath, actualProgram)

	return err
}
