/*
 *
 * RR2 - internal/constants/core.go
 *
 */

package constants

import "runtime"

const (
	VERSION_DATE = "2026-06-30"
	HIST_FILE    = ".ChippyHistory"

	RR_CONTEXT_DISPLAY_NAME     = "<ChippyProgram>"
	CLI_CONTEXT_DISPLAY_NAME_FN = "<ChippyREPL>"
	BASE_FUNC_NAME_FN_ANON      = "<ChippyAnon>"

	DIGITS         = "0123456789"
	LETTERS        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LETTERS_DIGITS = LETTERS + DIGITS

	FILE_EXT_PROG   = ".chp"
	FILE_EXT_HEADER = ".chh"

	STR_LPLVR = "1.0.23"
	STR_LPLCN = "pardalote"
	STR_LPLOS = runtime.GOOS
	STR_LPLAR = runtime.GOARCH

	STR_ERR = "__~err~__"
	STR_OK  = "__~ok~__"

	NUM_NUL = 0
	NUM_FAL = 0
	NUM_TRU = 1

	MAX_VALUE_DEPTH = 100000

	RT_ERROR_TITLE           = "Runtime Error"
	RT_ERROR_DEBUGGER_PREFIX = "Chippy Debugger (Trace):"
	E_INVALID_SYNTAX         = "Syntax Error"
	E_ILLEGAL_CHAR           = "Illegal Character"
	E_EXPECTED_CHAR          = "Expected Character"
	E_CYCLIC_VALUE           = "A value cannot contain itself"
	E_VALUE_TOO_DEEP         = "Value is nested too deeply to process"
)
