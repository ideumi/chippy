/*
 *
 * RR2 - internal/constants/core.go
 *
 */

package constants

import "runtime"

const (
	VERSION_DATE = "2026-02-26"
	HIST_FILE    = ".ChipLangHistory"

	RR_CONTEXT_DISPLAY_NAME     = "<ChipLangProgram>"
	CLI_CONTEXT_DISPLAY_NAME_FN = "<ChipLangREPL>"
	BASE_FUNC_NAME_FN_ANON      = "<ChipLangAnon>"

	DIGITS         = "0123456789"
	LETTERS        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LETTERS_DIGITS = LETTERS + DIGITS

	FILE_EXT_PROG   = ".chp"
	FILE_EXT_HEADER = ".chh"

	STR_LPLVR = "1.0.12"
	STR_LPLCN = "pardalote"
	STR_LPLOS = runtime.GOOS
	STR_LPLAR = runtime.GOARCH

	STR_ERR = "__~err~__"
	STR_OK  = "__~ok~__"

	NUM_NUL = 0
	NUM_FAL = 0
	NUM_TRU = 1

	RT_ERROR_TITLE           = "Runtime Error"
	RT_ERROR_DEBUGGER_PREFIX = "ChipLang Debugger (Trace):"
	E_INVALID_SYNTAX         = "Syntax Error"
	E_ILLEGAL_CHAR           = "Illegal Character"
	E_EXPECTED_CHAR          = "Expected Character"
)
