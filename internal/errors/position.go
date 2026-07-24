/*
 *
 * RR2 - internal/errors/position.go
 *
 */

package errors

import (
	"fmt"
	"unicode/utf8"
)

type Position struct {
	Index  int
	Line   int
	Column int
	File   string
	Text   string
}

func NewPosition(index, line, column int, file, text string) *Position {
	return &Position{
		Index:  index,
		Line:   line,
		Column: column,
		File:   file,
		Text:   text,
	}
}

func (p *Position) Advance(currentChar rune) *Position {
	p.Index += utf8.RuneLen(currentChar)
	p.Column++

	if currentChar == '\n' {
		p.Line++
		p.Column = 0
	}

	return p
}

func (p *Position) Copy() *Position {
	return &Position{
		Index:  p.Index,
		Line:   p.Line,
		Column: p.Column,
		File:   p.File,
		Text:   p.Text,
	}
}

func (p *Position) DisplayLine() int {
	return p.Line + 1
}

func (p *Position) String() string {
	return fmt.Sprintf("File %s, line %d", p.File, p.DisplayLine())
}
