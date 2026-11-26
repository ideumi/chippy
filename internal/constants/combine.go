/*
 *
 * Chippy - internal/constants/combine.go
 *
 */

package constants

const (
	// Combine tool defaults

	COMBINE_DEFAULT_FILENAME = "combine.chp"

	// Combine subcommands

	COMBINE_NEW_COMMAND  = "new"
	COMBINE_RUN_COMMAND  = "run"
	COMBINE_TEST_COMMAND = "test"

	// Combine test runner

	COMBINE_TEST_FILE_SUFFIX     = "-test" + FILE_EXT_PROG
	COMBINE_TEST_FUNCTION_PREFIX = "Test"

	// File permissions

	FILE_PERM_READABLE   = 0644
	FILE_PERM_EXECUTABLE = 0755

	// Config variable names

	CONFIG_PROJECT          = "Project"
	CONFIG_VERSION          = "Version"
	CONFIG_LICENCE          = "Licence"
	CONFIG_OUTPUT           = "Output"
	CONFIG_SOURCE           = "Source"
	CONFIG_PATHS            = "Paths"
	CONFIG_STRIP_COMMENTS   = "StripComments"
	CONFIG_STRIP_WHITESPACE = "StripWhitespace"
	CONFIG_ADD_SHEBANG      = "AddShebang"
	CONFIG_EXTERNAL         = "External"
	CONFIG_COMPILE          = "Compile"

	// Failure reasons

	COMBINE_E_BUNDLE_ERRORS     = "cannot bundle due to errors"
	COMBINE_E_BUNDLE_COLLISIONS = "cannot bundle due to symbol collisions"
	COMBINE_E_RUN_ERRORS        = "cannot run due to errors"
	COMBINE_E_RUN_COLLISIONS    = "cannot run due to symbol collisions"
)
