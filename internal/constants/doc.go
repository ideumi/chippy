/*
 *
 * RR2 - internal/constants/doc.go
 *
 */

package constants

const (
	// Documentation paths

	DOC_DIR_LOCAL             = "doc"
	DOC_DIR_USER              = ".local/share/chippy/doc"
	DOC_DIR_PRODUCTION        = "/usr/share/doc/chippy"
	DOC_DIR_PRODUCTION_TERMUX = "/data/data/com.termux/files/usr/share/doc/chippy"
	DOC_FILE_EXTENSION        = ".chpdoc"

	// Renderer constants

	CODE_BLOCK_DELIMITER = "'''"
	INLINE_CODE_MARKER   = "''"
	BOLD_MARKER          = "*"

	CODE_LINE_START = 1
	COLUMN_WIDTH    = 20
	NUM_COLUMNS     = 3

	// Symbol documentation constants

	SYMBOL_DECL_OPEN  = "<"
	SYMBOL_DECL_CLOSE = ">"

	SYMBOL_FILTER_ALL = "all"
)
