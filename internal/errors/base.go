/*
 *
 * RR2 - internal/errors/base.go
 *
 */

package errors

import (
	"fmt"
	"strings"
)

type BaseError struct {
	PosStart  *Position
	PosEnd    *Position
	ErrorName string
	Details   string
}

func NewBaseError(posStart, posEnd *Position, errorName, details string) *BaseError {
	return &BaseError{
		PosStart:  posStart,
		PosEnd:    posEnd,
		ErrorName: errorName,
		Details:   details,
	}
}

func (e *BaseError) AsString() string {
	result := fmt.Sprintf("%s: %s", e.ErrorName, e.Details)

	if e.PosStart != nil {
		result += fmt.Sprintf("\nFile %s, line %d", e.PosStart.File, e.PosStart.Line+1)

		if e.PosStart.Text != "" {
			result += "\n\n" + e.stringWithArrows()
		}
	}

	return result
}

func (e *BaseError) Error() string {
	return e.AsString()
}

func (e *BaseError) stringWithArrows() string {
	if e.PosStart == nil {
		return ""
	}

	result := ""
	text := e.PosStart.Text

	idxStart := max(0, strings.LastIndex(text[:e.PosStart.Index], "\n"))

	if idxStart > 0 {
		idxStart += 1
	}

	idxEnd := strings.Index(text[e.PosStart.Index:], "\n")

	if idxEnd == -1 {
		idxEnd = len(text)
	} else {
		idxEnd += e.PosStart.Index
	}

	lineCount := strings.Count(text[:idxStart], "\n")

	line := text[idxStart:idxEnd]
	// Replace tabs with 4 spaces for consistent display
	line = strings.ReplaceAll(line, "\t", "    ")

	// Calculate relative column positions within the extracted line
	colStart := e.PosStart.Index - idxStart
	colEnd := colStart

	if e.PosEnd != nil {
		colEnd = e.PosEnd.Index - idxStart
	}

	result += fmt.Sprintf("Line %d: %s\n", lineCount+1, line)

	// Convert character position to visual position accounting for tab replacement
	visualColStart := 0
	for i := 0; i < colStart && i < len(text[idxStart:idxEnd]); i++ {
		if text[idxStart+i] == '\t' {
			visualColStart += 4 // Each tab becomes 4 spaces
		} else {
			visualColStart++
		}
	}

	if visualColStart < 0 {
		visualColStart = 0
	}

	arrows := strings.Repeat(" ", len(fmt.Sprintf("Line %d: ", lineCount+1))+visualColStart)

	if e.PosEnd != nil && colEnd > colStart {
		// Calculate visual width of the error span
		visualColEnd := 0
		for i := 0; i < colEnd && i < len(text[idxStart:idxEnd]); i++ {
			if text[idxStart+i] == '\t' {
				visualColEnd += 4 // Each tab becomes 4 spaces
			} else {
				visualColEnd++
			}
		}
		arrows += strings.Repeat("^", visualColEnd-visualColStart)
	} else {
		arrows += "^"
	}

	result += arrows

	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
