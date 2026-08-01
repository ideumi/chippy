/*
 *
 * Chippy - internal/constants/core.go
 *
 */

package constants

import "runtime"

const (
	VERSION_DATE = "2026-07-31"
	HIST_FILE    = ".ChippyHistory"

	CONTEXT_DISPLAY_NAME        = "<ChippyProgram>"
	CLI_CONTEXT_DISPLAY_NAME_FN = "<ChippyREPL>"
	BASE_FUNC_NAME_FN_ANON      = "<ChippyAnon>"

	DIGITS         = "0123456789"
	LETTERS        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LETTERS_DIGITS = LETTERS + DIGITS

	FILE_EXT_PROG   = ".chp"
	FILE_EXT_HEADER = ".chh"

	CHIPPY_SHEBANG = "#!/usr/bin/chippy"

	STR_LPLVR = "1.1.0"
	STR_LPLCN = "rixosa"
	STR_LPLOS = runtime.GOOS
	STR_LPLAR = runtime.GOARCH

	STR_ERR = "__~err~__"
	STR_OK  = "__~ok~__"

	NUM_NUL = 0
	NUM_FAL = 0
	NUM_TRU = 1

	RT_ERROR_TITLE      = "Runtime Error"
	PANIC_ERROR_TITLE   = "Panic"
	DAMAGED_ERROR_TITLE = "Damaged Program"
	E_INVALID_SYNTAX    = "Syntax Error"
	E_ILLEGAL_CHAR      = "Illegal Character"
	E_EXPECTED_CHAR     = "Expected Character"
	E_CYCLIC_VALUE      = "A value cannot contain itself"
	E_VALUE_TOO_DEEP    = "Value is nested too deeply"
	E_DAMAGED_PROGRAM   = "This program is damaged"
)
