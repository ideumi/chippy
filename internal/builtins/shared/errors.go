/*
 *
 * RR2 - internal/builtins/shared/errors.go
 *
 */

package shared

import "fmt"

// BuiltinErrors provides centralized error message generation for builtin functions
type BuiltinErrors struct{}

// Global instance for all builtins to use
var Errors = BuiltinErrors{}

// InvalidArgCount generates error for wrong number of arguments
func (e BuiltinErrors) InvalidArgCount(funcName string, expected int) string {
	if expected == 0 {
		return fmt.Sprintf("%s() takes no arguments", funcName)

	} else if expected == 1 {
		return fmt.Sprintf("%s() takes exactly one argument", funcName)

	} else {
		return fmt.Sprintf("%s() takes exactly %d arguments", funcName, expected)
	}
}

// InvalidArgCountWithHint generates error for wrong number of arguments with parameter hints
func (e BuiltinErrors) InvalidArgCountWithHint(funcName string, expected int, hint string) string {
	if expected == 0 {
		return fmt.Sprintf("%s() takes no arguments", funcName)

	} else if expected == 1 {
		return fmt.Sprintf("%s() takes exactly 1 argument (%s)", funcName, hint)

	} else {
		return fmt.Sprintf("%s() takes exactly %d arguments (%s)", funcName, expected, hint)
	}
}

// InvalidArgType generates error for wrong argument type
func (e BuiltinErrors) InvalidArgType(funcName string, expectedType string) string {
	return fmt.Sprintf("%s() argument must be %s", funcName, expectedType)
}

// InvalidArgTypePositional generates error for wrong argument type at specific position
func (e BuiltinErrors) InvalidArgTypePositional(funcName string, position string, expectedType string) string {
	return fmt.Sprintf("%s() %s argument must be %s", funcName, position, expectedType)
}

// InvalidArgTypeWithHint generates error for wrong argument type with descriptive hint
func (e BuiltinErrors) InvalidArgTypeWithHint(funcName string, expectedType, hint string) string {
	return fmt.Sprintf("%s() argument must be %s (%s)", funcName, expectedType, hint)
}

// InvalidArgTypePositionalWithHint generates error for wrong argument type at position with hint
func (e BuiltinErrors) InvalidArgTypePositionalWithHint(funcName string, position, expectedType, hint string) string {
	return fmt.Sprintf("%s() %s argument must be %s (%s)", funcName, position, expectedType, hint)
}

// InvalidValue generates error for value constraint violations
func (e BuiltinErrors) InvalidValue(constraint string) string {
	return constraint
}

// CannotConvert generates error for impossible conversions
func (e BuiltinErrors) CannotConvert(what, reason string) string {
	return fmt.Sprintf("Cannot convert %s%s", what, reason)
}

// Common type names
const (
	TypeNumber       = "a number"
	TypeString       = "a string"
	TypeList         = "a list"
	TypeBytes        = "bytes"
	TypeListOrBytes  = "list or bytes"
	TypeStringOrList = "string or list"
	TypeFunction     = "a function"
)

// Common position names
const (
	PositionFirst  = "first"
	PositionSecond = "second"
	PositionThird  = "third"
	PositionFourth = "fourth"
)
