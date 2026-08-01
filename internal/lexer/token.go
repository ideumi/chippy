/*
 *
 * Chippy - internal/lexer/token.go
 *
 */

package lexer

import (
	"chip-go/internal/errors"
	"fmt"
)

type Token struct {
	Type     string
	Value    interface{}
	PosStart *errors.Position
	PosEnd   *errors.Position
}

func NewToken(tokenType string, value interface{}, posStart, posEnd *errors.Position) *Token {
	token := &Token{
		Type:     tokenType,
		Value:    value,
		PosStart: posStart,
	}

	if posEnd != nil {
		token.PosEnd = posEnd
	} else if posStart != nil {
		token.PosEnd = posStart.Copy()
		token.PosEnd.Advance(0) // Move to next position
	}

	return token
}

func (t *Token) Matches(tokenType string, value interface{}) bool {
	return t.Type == tokenType && t.Value == value
}

func (t *Token) String() string {
	if t.Value != nil {
		return fmt.Sprintf("%s:%v", t.Type, t.Value)
	}
	return t.Type
}
