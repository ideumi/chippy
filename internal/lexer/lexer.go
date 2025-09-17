/*
 *
 * RR2 - internal/lexer/lexer.go
 *
 */

package lexer

import (
	"strconv"
	"unicode"
	"unicode/utf8"

	"chip-go/internal/constants"
	"chip-go/internal/errors"
)

type Lexer struct {
	file        string
	text        string
	pos         *errors.Position
	currentChar rune
	byteIndex   int
}

func NewLexer(file, text string) *Lexer {
	lexer := &Lexer{
		file:      file,
		text:      text,
		pos:       errors.NewPosition(-1, 0, -1, file, text),
		byteIndex: -1,
	}
	lexer.advance()

	return lexer
}

func (l *Lexer) advance() {
	l.pos.Advance(l.currentChar)
	l.byteIndex++
	if l.byteIndex < len(l.text) {
		r, size := utf8.DecodeRuneInString(l.text[l.byteIndex:])
		l.currentChar = r
		l.byteIndex += size - 1 // -1 because we already incremented byteIndex
	} else {
		l.currentChar = 0
	}
}

func (l *Lexer) MakeTokens() ([]*Token, error) {
	var tokens []*Token

	for l.currentChar != 0 {
		switch {

		case l.currentChar == ' ' || l.currentChar == '\t':
			l.advance()

		case l.currentChar == '#':
			l.skipComment()

		case l.currentChar == '\n' || l.currentChar == '\r':
			tokens = append(tokens, NewToken(constants.TT_NEWLINE, nil, l.pos.Copy(), nil))
			l.advance()

		case unicode.IsDigit(l.currentChar):
			token, err := l.makeNumber()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)

		case unicode.IsLetter(l.currentChar) || l.currentChar == '_':
			token := l.makeIdentifier()
			// Check for byte literal syntax 'b['
			if token.Type == constants.TT_IDENTIFIER && token.Value == "b" && l.currentChar == '[' {
				token.Type = constants.TT_BYTELITERAL
			}
			tokens = append(tokens, token)

		case l.currentChar == '"':
			token, err := l.makeString()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)

		case l.currentChar == '+':
			tokens = append(tokens, NewToken(constants.TT_PLUS, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '-':
			token := l.makeMinus()
			tokens = append(tokens, token)

		case l.currentChar == '*':
			tokens = append(tokens, NewToken(constants.TT_MUL, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '/':
			tokens = append(tokens, NewToken(constants.TT_DIV, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '%':
			tokens = append(tokens, NewToken(constants.TT_MOD, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '^':
			tokens = append(tokens, NewToken(constants.TT_POW, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '(':
			tokens = append(tokens, NewToken(constants.TT_LPAREN, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == ')':
			tokens = append(tokens, NewToken(constants.TT_RPAREN, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '[':
			tokens = append(tokens, NewToken(constants.TT_LSQUARE, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == ']':
			tokens = append(tokens, NewToken(constants.TT_RSQUARE, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '{':
			tokens = append(tokens, NewToken(constants.TT_LBRACE, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '}':
			tokens = append(tokens, NewToken(constants.TT_RBRACE, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == ',':
			tokens = append(tokens, NewToken(constants.TT_COMMA, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == ';':
			tokens = append(tokens, NewToken(constants.TT_SEMICOLON, nil, l.pos.Copy(), nil))
			l.advance()

		case l.currentChar == '!':
			token, err := l.makeNotEquals()
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)

		case l.currentChar == '=':
			tokens = append(tokens, l.makeEquals())

		case l.currentChar == '<':
			tokens = append(tokens, l.makeLessThan())

		case l.currentChar == '>':
			tokens = append(tokens, l.makeGreaterThan())

		default:
			posStart := l.pos.Copy()
			char := l.currentChar
			l.advance()

			return nil, errors.NewIllegalCharError(posStart, l.pos.Copy(), "'"+string(char)+"'")
		}
	}

	tokens = append(tokens, NewToken(constants.TT_EOF, nil, l.pos.Copy(), nil))

	return tokens, nil
}

func (l *Lexer) skipComment() {
	l.advance()
	for l.currentChar != 0 && l.currentChar != '\n' {
		l.advance()
	}
}

func (l *Lexer) makeNumber() (*Token, error) {
	numStr := ""
	dotCount := 0
	posStart := l.pos.Copy()

	for l.currentChar != 0 && (unicode.IsDigit(l.currentChar) || l.currentChar == '.') {
		if l.currentChar == '.' {
			if dotCount == 1 {
				break
			}
			dotCount++
		}
		numStr += string(l.currentChar)
		l.advance()
	}

	if dotCount == 0 {
		val, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return nil, errors.NewIllegalCharError(posStart, l.pos.Copy(), "Invalid integer: "+numStr)
		}

		return NewToken(constants.TT_INT, val, posStart, l.pos.Copy()), nil
	} else {
		val, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return nil, errors.NewIllegalCharError(posStart, l.pos.Copy(), "Invalid float: "+numStr)
		}

		return NewToken(constants.TT_FLOAT, val, posStart, l.pos.Copy()), nil
	}
}

func (l *Lexer) makeIdentifier() *Token {
	idStr := ""
	posStart := l.pos.Copy()

	for l.currentChar != 0 && (unicode.IsLetter(l.currentChar) || unicode.IsDigit(l.currentChar) || l.currentChar == '_') {
		idStr += string(l.currentChar)
		l.advance()
	}

	if constants.IsKeyword(idStr) {
		return NewToken(constants.TT_KEYWORD, idStr, posStart, l.pos.Copy())
	} else {
		return NewToken(constants.TT_IDENTIFIER, idStr, posStart, l.pos.Copy())
	}
}

func (l *Lexer) makeString() (*Token, error) {
	str := ""
	posStart := l.pos.Copy()
	l.advance() // Skip opening quote

	escapeChars := map[rune]rune{
		'n':  '\n',
		't':  '\t',
		'r':  '\r',
		'\\': '\\',
		'"':  '"',
	}

	for l.currentChar != 0 && l.currentChar != '"' {
		if l.currentChar == '\\' {
			l.advance()
			if escapeChar, exists := escapeChars[l.currentChar]; exists {
				str += string(escapeChar)
			} else {
				str += string(l.currentChar)
			}
		} else {
			str += string(l.currentChar)
		}

		l.advance()
	}

	if l.currentChar != '"' {
		return nil, errors.NewExpectedCharError(posStart, l.pos.Copy(), "'\"'")
	}

	l.advance() // Skip closing quote

	return NewToken(constants.TT_STRING, str, posStart, l.pos.Copy()), nil
}

func (l *Lexer) makeMinus() *Token {
	posStart := l.pos.Copy()
	l.advance()

	return NewToken(constants.TT_MINUS, nil, posStart, l.pos.Copy())
}

func (l *Lexer) makeNotEquals() (*Token, error) {
	posStart := l.pos.Copy()
	l.advance()

	if l.currentChar == '=' {
		l.advance()
		return NewToken(constants.TT_NE, nil, posStart, l.pos.Copy()), nil
	}

	return nil, errors.NewExpectedCharError(posStart, l.pos.Copy(), "'=' (after '!')")
}

func (l *Lexer) makeEquals() *Token {
	posStart := l.pos.Copy()
	l.advance()

	if l.currentChar == '=' {
		l.advance()
		return NewToken(constants.TT_EE, nil, posStart, l.pos.Copy())
	}

	return NewToken(constants.TT_EQ, nil, posStart, l.pos.Copy())
}

func (l *Lexer) makeLessThan() *Token {
	posStart := l.pos.Copy()
	l.advance()

	if l.currentChar == '=' {
		l.advance()
		return NewToken(constants.TT_LTE, nil, posStart, l.pos.Copy())
	}

	return NewToken(constants.TT_LT, nil, posStart, l.pos.Copy())
}

func (l *Lexer) makeGreaterThan() *Token {
	posStart := l.pos.Copy()
	l.advance()

	if l.currentChar == '=' {
		l.advance()
		return NewToken(constants.TT_GTE, nil, posStart, l.pos.Copy())
	}

	return NewToken(constants.TT_GT, nil, posStart, l.pos.Copy())
}
