/*
 *
 * Combine Test Runner
 *
 */

package combine

import (
	"chip-go/internal/bytecode"
	"chip-go/internal/bytecode/serializer"
	"chip-go/internal/compiler"
	"chip-go/internal/constants"
	"chip-go/internal/handles"
	"chip-go/internal/modena"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	testColorPass = "\033[1;32m"
	testColorFail = "\033[1;31m"
	testColorBold = "\033[1m"
)

type testResult struct {
	name     string
	pass     bool
	duration time.Duration
	output   string
	err      error
}

type suiteFile struct {
	rel           string
	abs           string
	program       *resolvedProgram
	names         []string
	selectedNames []string
	buildFailed   bool
	buildErr      error
	parseErrs     []ValidationError
	collisions    []SymbolCollision
}

type fileResults struct {
	file  suiteFile
	tests []testResult
}

// Test runs a combine project's test suite in memory
func Test(cfgFile string, rest []string) error {
	config, err := loadCombineConfig(cfgFile)

	if err != nil {
		return err
	}

	fileSlot, pattern, parseErr := parseTestQuery(rest)

	if parseErr != nil {
		return parseErr
	}

	configDir := filepath.Clean(filepath.Dir(cfgFile))

	testFiles, err := findTestFiles(configDir, cfgFile)

	if err != nil {
		return err
	}

	if fileSlot != "" {
		testFiles = filterTestFiles(testFiles, fileSlot)

		if len(testFiles) == 0 {
			return fmt.Errorf("test file '%s' not found", fileSlot)
		}
	}

	if len(testFiles) == 0 {
		return errors.New("no test files found")
	}

	return runTestSuite(config, configDir, fileSlot, testFiles, pattern)
}

func parseTestQuery(rest []string) (fileSlot string, pattern string, err error) {
	if len(rest) == 0 {
		return "", "", nil
	}

	if isTestFileArg(rest[0]) {
		if len(rest) > 2 {
			return "", "", errors.New("too many arguments")
		}

		fileSlot = rest[0]

		if len(rest) == 2 {
			pattern = rest[1]
		}

		return fileSlot, pattern, nil
	}

	if len(rest) > 1 {
		return "", "", errors.New("expected at most one pattern")
	}

	return "", rest[0], nil
}

func isTestFileArg(arg string) bool {
	return strings.Contains(arg, "/") || strings.HasSuffix(arg, constants.FILE_EXT_PROG)
}

func findTestFiles(root string, cfgFile string) ([]string, error) {
	var found []string

	cfgAbs, cfgErr := filepath.Abs(cfgFile)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != root && strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}

			return nil
		}

		// The config file is not a test file, even when it carries the suffix.
		if cfgErr == nil {
			if abs, absErr := filepath.Abs(path); absErr == nil && abs == cfgAbs {
				return nil
			}
		}

		if !strings.HasSuffix(path, constants.COMBINE_TEST_FILE_SUFFIX) {
			return nil
		}

		data, readErr := os.ReadFile(path)

		if readErr != nil {
			return readErr
		}

		if serializer.LooksLikeBytecode(data) {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)

		if relErr != nil {
			return relErr
		}

		found = append(found, rel)

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(found)

	return found, nil
}

func filterTestFiles(files []string, fileSlot string) []string {
	want := filepath.Clean(fileSlot)

	var out []string

	for _, file := range files {
		if filepath.Clean(file) == want {
			out = append(out, file)
		}
	}

	return out
}

func runTestSuite(config CombineConfig, configDir string, fileSlot string, testFiles []string, pattern string) error {
	files := buildTestFiles(configDir, testFiles, config.Paths, config.External)

	matched := selectTestFunctions(allTestNames(files), pattern)

	if pattern != "" && len(matched) == 0 {
		// A file that failed to build may still hold a test matching the pattern.
		// Its names are unknown, so show the failure instead of reporting no match.
		if !anyBuildFailed(files) {
			return fmt.Errorf("no test matches pattern '%s'", pattern)
		}
	}

	matchedSet := make(map[string]bool)

	for _, name := range matched {
		matchedSet[name] = true
	}

	files = selectTestFiles(files, matchedSet, pattern)

	if len(files) == 0 {
		if fileSlot != "" {
			return fmt.Errorf("no tests in file '%s'", fileSlot)
		}

		return errors.New("no tests found")
	}

	results, hasFailure := executeTestFiles(files)

	printTestReport(config, results)

	if hasFailure {
		os.Exit(1)
	}

	return nil
}

func buildTestFiles(configDir string, testFiles []string, paths []string, external []string) []suiteFile {
	var files []suiteFile

	for _, rel := range testFiles {
		files = append(files, buildTestFile(configDir, rel, paths, external))
	}

	return files
}

func buildTestFile(configDir, rel string, paths []string, external []string) suiteFile {
	file := suiteFile{rel: rel, abs: filepath.Join(configDir, rel)}

	program, err := resolveProgram(file.abs, paths, external)

	if err != nil {
		file.buildErr = err
		file.buildFailed = true

		return file
	}

	if len(program.parseErrs) > 0 {
		file.parseErrs = program.parseErrs
		file.buildFailed = true

		return file
	}

	if len(program.collisions) > 0 {
		file.collisions = program.collisions
		file.buildFailed = true

		return file
	}

	// Test names come from the test file itself, not its dependencies.
	chunk, err := compiler.CompileSource(file.abs, program.entry.source)

	if err != nil {
		file.buildErr = err
		file.buildFailed = true

		return file
	}

	file.program = program
	file.names = testFunctionNames(chunk)

	return file
}

func testFunctionNames(chunk *bytecode.Chunk) []string {
	var names []string

	for _, fn := range chunk.Functions {
		if strings.HasPrefix(fn.Name, constants.COMBINE_TEST_FUNCTION_PREFIX) {
			names = append(names, fn.Name)
		}
	}

	return names
}

func allTestNames(files []suiteFile) []string {
	var names []string

	for _, file := range files {
		names = append(names, file.names...)
	}

	return names
}

func anyBuildFailed(files []suiteFile) bool {
	for _, file := range files {
		if file.buildFailed {
			return true
		}
	}

	return false
}

// testPassed applies the result rule: PASS only when the call returned zero.
// aka. return 0; or fall off
func testPassed(value values.Value) bool {
	if !value.IsInt() {
		return false
	}

	code, _ := value.AsInt()

	return code == 0
}

func testTier(name, query string) int {
	switch {
	case name == query:
		return 3
	case strings.HasPrefix(name, query):
		return 2
	case strings.Contains(name, query):
		return 1
	}

	return 0
}

// This is more or less the same logic that is used in doc.go for matching
func selectTestFunctions(names []string, pattern string) []string {
	query := strings.ToLower(pattern)

	if query == "" {
		return names
	}

	bestTier := 0

	for _, name := range names {
		if tier := testTier(strings.ToLower(name), query); tier > bestTier {
			bestTier = tier
		}
	}

	if bestTier == 0 {
		return nil
	}

	var matched []string
	seen := make(map[string]bool)

	for _, name := range names {
		if testTier(strings.ToLower(name), query) == bestTier && !seen[name] {
			seen[name] = true
			matched = append(matched, name)
		}
	}

	return matched
}

// selectTestFiles stores each file's selected tests on the suiteFile and drops
// the files with none.
// This keeps files that failed to build, regardless if selected or not.
func selectTestFiles(files []suiteFile, matchedSet map[string]bool, pattern string) []suiteFile {
	var shown []suiteFile

	for _, file := range files {
		if file.buildFailed {
			shown = append(shown, file)

			continue
		}

		for _, name := range file.names {
			if pattern == "" || matchedSet[name] {
				file.selectedNames = append(file.selectedNames, name)
			}
		}

		if len(file.selectedNames) > 0 {
			shown = append(shown, file)
		}
	}

	return shown
}

func executeTestFiles(files []suiteFile) ([]fileResults, bool) {
	var results []fileResults
	hasFailure := false

	for _, file := range files {
		var fileTests []testResult

		for _, name := range file.selectedNames {
			pass, output, testErr, duration := runOneTest(file.program, name)

			fileTests = append(fileTests, testResult{name: name, pass: pass, duration: duration, output: output, err: testErr})

			if !pass {
				hasFailure = true
			}
		}

		if file.buildFailed {
			hasFailure = true
		}

		results = append(results, fileResults{file: file, tests: fileTests})
	}

	return results, hasFailure
}

// runOneTest runs one test function in a new VM. It returns pass, the captured
// output, the failure reason (nil when passed), and the time it took to run the
// test.
func runOneTest(program *resolvedProgram, name string) (bool, string, error, time.Duration) {
	start := time.Now()

	mod := modena.New()

	// Capture STDOUT and STDERR to show in test output
	read, write, pipeErr := os.Pipe()

	var output string
	var outCh chan []byte

	if pipeErr == nil {
		outCh = make(chan []byte, 1)

		go func() {
			data, _ := io.ReadAll(read)

			outCh <- data
		}()

		// Set VM standard handles to pipe, I can't stop drinking oil, forgive me
		// VM gods
		vmFiles(mod).Store(1, write)
		vmFiles(mod).Store(2, write)
	}

	entryValue, entryHalted, retErr := executeResolved(program, mod, nil)

	pass := false
	value := values.Value{}

	if retErr == nil && entryHalted {
		code, _ := entryValue.AsInt()

		retErr = fmt.Errorf("exited with %d", code)
	}

	if retErr == nil {
		value, retErr = mod.CallFunction(name)
	}

	if pipeErr == nil {
		// Restore the VM and close write so the reader terminates and the captured
		// output can be collected.
		vmFiles(mod).Store(1, os.Stdout)
		vmFiles(mod).Store(2, os.Stderr)

		write.Close()

		output = string(<-outCh)

		read.Close()
	}

	if retErr == nil && testPassed(value) {
		pass = true
	} else if retErr == nil {
		retErr = fmt.Errorf("returned %s", value.String())
	}

	elapsed := time.Since(start)

	if elapsed < time.Millisecond {
		elapsed = time.Millisecond
	}

	return pass, output, retErr, elapsed
}

func vmFiles(mod *modena.Modena) *handles.FileHandles {
	return orchestrator.Get().GetRegistry(mod.GetGlobalContext().InstanceID).Files
}

func printTestReport(config CombineConfig, results []fileResults) {
	maxNameLen, passedTests, failedTests := 0, 0, 0

	var total time.Duration

	for _, entry := range results {
		file := entry.file

		if file.buildFailed {
			failedTests++
		}

		for _, test := range entry.tests {
			if len(test.name) > maxNameLen {
				maxNameLen = len(test.name)
			}

			total += test.duration

			if test.pass {
				passedTests++
			} else {
				failedTests++
			}
		}
	}

	fmt.Printf("Testing %s V-%s:\n", config.Project, config.Version)

	for _, entry := range results {
		file := entry.file

		if file.buildFailed {
			fmt.Printf("  %s  %s\n", testColorize("FAIL", testColorFail), testColorize(file.rel, testColorBold))

			if len(file.parseErrs) > 0 {
				fmt.Println(indentReportBlock(parseErrorsText(file.parseErrs)))

				fmt.Println()
			}

			if len(file.collisions) > 0 {
				fmt.Println(indentReportBlock(collisionsText(file.collisions)))

				fmt.Println()
			}

			if len(file.parseErrs) == 0 && len(file.collisions) == 0 {
				fmt.Println(indentReportBlock(file.buildErr.Error()))

				fmt.Println()
			}

			continue
		}

		fileFailed := false

		for _, test := range entry.tests {
			if !test.pass {
				fileFailed = true
			}
		}

		status, code := "PASS", testColorPass

		if fileFailed {
			status, code = "FAIL", testColorFail
		}

		fmt.Printf("  %s  %s\n", testColorize(status, code), testColorize(file.rel, testColorBold))

		for _, test := range entry.tests {
			status, code := "PASS", testColorPass

			if !test.pass {
				status, code = "FAIL", testColorFail
			}

			line := fmt.Sprintf("    %s  %-*s  %s", testColorize(status, code), maxNameLen, test.name, test.duration.Truncate(time.Millisecond).String())

			multiLine := test.err != nil && strings.Contains(test.err.Error(), "\n")

			if test.err != nil && !multiLine {
				line += "  " + test.err.Error()
			}

			fmt.Println(line)

			if test.output != "" {
				for _, outLine := range strings.Split(strings.TrimSuffix(test.output, "\n"), "\n") {
					fmt.Printf("    %s\n", outLine)
				}
			}

			if multiLine {
				fmt.Println(indentReportBlock(test.err.Error()))
			}
		}

		fmt.Println()
	}

	fmt.Printf("Total: %d file(s), %d passed, %d failed, %s\n", len(results), passedTests, failedTests, total.Truncate(time.Millisecond).String())
}

// indentReportBlock aligns multi-line error blocks to the report's columns.
func indentReportBlock(block string) string {
	lines := strings.Split(block, "\n")

	for i, line := range lines {
		lines[i] = "    " + line
	}

	return strings.Join(lines, "\n")
}

func testColorize(s, code string) string {
	return code + s + "\033[0m"
}
