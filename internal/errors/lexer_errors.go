/*
 *
 * RR2 - internal/errors/lexer_errors.go
 *
 */

package errors

import "chip-go/internal/constants"

type IllegalCharError struct {
	*BaseError
}

func NewIllegalCharError(posStart, posEnd *Position, details string) *IllegalCharError {
	return &IllegalCharError{
		BaseError: NewBaseError(posStart, posEnd, constants.E_ILLEGAL_CHAR, details),
	}
}

type ExpectedCharError struct {
	*BaseError
}

func NewExpectedCharError(posStart, posEnd *Position, details string) *ExpectedCharError {
	return &ExpectedCharError{
		BaseError: NewBaseError(posStart, posEnd, constants.E_EXPECTED_CHAR, details),
	}
}

type InvalidSyntaxError struct {
	*BaseError
}

func NewInvalidSyntaxError(posStart, posEnd *Position, details string) *InvalidSyntaxError {
	return &InvalidSyntaxError{
		BaseError: NewBaseError(posStart, posEnd, constants.E_INVALID_SYNTAX, details),
	}
}

type RTError struct {
	*BaseError
}

func NewRTError(posStart, posEnd *Position, details string) *RTError {
	return &RTError{
		BaseError: NewBaseError(posStart, posEnd, constants.RT_ERROR_TITLE, details),
	}
}
