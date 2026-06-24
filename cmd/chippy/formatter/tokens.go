/*
 *
 * The Chippy formatter engine / tokens part
 *
 */

package formatter

import (
	"chip-go/internal/constants"
	"chip-go/internal/lexer"
)

func isElse(t *lexer.Token) bool {
	return t.Type == constants.TT_KEYWORD && (t.Value == "else" || t.Value == "elseif")
}

func isCommentTok(t *lexer.Token) bool {
	return t.Type == constants.TT_COMMENT || t.Type == constants.TT_DOC_COMMENT
}

// isTerminator reports whether t ends a statement. A nil t (start of file) counts.
func isTerminator(t *lexer.Token) bool {
	if t == nil {
		return true
	}

	switch t.Type {
	case constants.TT_SEMICOLON, constants.TT_LBRACE, constants.TT_RBRACE:
		return true
	}

	return false
}

func isCloser(t *lexer.Token) bool {
	switch t.Type {
	case constants.TT_RBRACE, constants.TT_RPAREN, constants.TT_RSQUARE:
		return true
	}

	return false
}

// isValueEnd reports whether t ends an operand.
func isValueEnd(t *lexer.Token) bool {
	if t == nil {
		return false
	}

	switch t.Type {
	case constants.TT_INT, constants.TT_FLOAT, constants.TT_STRING,
		constants.TT_IDENTIFIER, constants.TT_RPAREN, constants.TT_RSQUARE:
		return true
	}

	return false
}

// wantSpace reports whether a space belongs between prev and cur on one line.
func wantSpace(prev, cur *lexer.Token, prevUnaryMinus bool) bool {
	// Empty block stays compact: '{}'.
	if prev.Type == constants.TT_LBRACE && cur.Type == constants.TT_RBRACE {
		return false
	}

	switch cur.Type {
	case constants.TT_RPAREN, constants.TT_RSQUARE,
		constants.TT_COMMA, constants.TT_SEMICOLON, constants.TT_COLON:
		return false
	}

	switch prev.Type {
	case constants.TT_LPAREN, constants.TT_LSQUARE:
		return false
	}

	if prevUnaryMinus {
		return false
	}

	// '(' / '[' cuddle as a call/index after a value (b[...] and m[...] literals
	// included) or a 'func' parameter list. After any other keyword or operator,
	// space.
	if cur.Type == constants.TT_LPAREN || cur.Type == constants.TT_LSQUARE {
		switch prev.Type {
		case constants.TT_IDENTIFIER, constants.TT_RPAREN, constants.TT_RSQUARE,
			constants.TT_RBRACE, constants.TT_INT, constants.TT_FLOAT, constants.TT_STRING,
			constants.TT_BYTES, constants.TT_MAP:
			return false
		case constants.TT_KEYWORD:
			if prev.Value == "func" {
				return false
			}
		}

		return true
	}

	return true
}
