/*
 *
 * The Chippy formatter engine / layout part
 *
 */

package formatter

import (
	"chip-go/internal/constants"
	"chip-go/internal/lexer"
	"strings"
)

func formatTokens(src string, tokens []*lexer.Token) string {
	inline := inlineBlocks(src, tokens)

	var b strings.Builder
	b.Grow(len(src)) // output is roughly the size of the input

	braceDepth := 0
	groupDepth := 0      // open '(' and '['
	pendingNewlines := 0 // buffered TT_NEWLINE since the last emitted token
	commentSincePrevSig := false
	prevUnaryMinus := false
	started := false

	var prevSig *lexer.Token // last emitted code token
	var blockStack []bool    // inline flag for each open '{'

	text := func(t *lexer.Token) string {
		s := src[t.PosStart.Index:t.PosEnd.Index]

		// Trailing whitespace inside a comment is insignificant
		if isCommentTok(t) {
			s = strings.TrimRight(s, " \t")
		}

		return s
	}

	writeLineStart := func(tok *lexer.Token, capN int) {
		n := pendingNewlines

		if n > capN {
			n = capN
		}

		for i := 0; i < n; i++ {
			b.WriteString("\n")
		}

		pendingNewlines = 0

		indent := braceDepth + groupDepth

		// A wrapped statement gets one extra level. Comments, closers,
		// and the '{'/else/elseif that begin a block do not. Mid-statement
		// is read from prevSig, the last code token, so a comment in the
		// middle does not hide it.
		if !isCommentTok(tok) && !isCloser(tok) && groupDepth == 0 &&
			tok.Type != constants.TT_LBRACE && !isElse(tok) &&
			!isTerminator(prevSig) {
			indent++
		}

		if indent < 0 {
			indent = 0
		}

		b.WriteString(strings.Repeat("\t", indent))
	}

	for idx, tok := range tokens {
		if tok.Type == constants.TT_EOF {
			continue
		}

		if tok.Type == constants.TT_NEWLINE {
			if started {
				pendingNewlines++
			}

			continue
		}

		// Closers dedent the line they begin.
		switch tok.Type {
		case constants.TT_RBRACE:
			braceDepth--

			if braceDepth < 0 {
				braceDepth = 0
			}
		case constants.TT_RPAREN, constants.TT_RSQUARE:
			groupDepth--

			if groupDepth < 0 {
				groupDepth = 0
			}
		}

		isComment := isCommentTok(tok)
		parentInline := len(blockStack) > 0 && blockStack[len(blockStack)-1]

		closingInline := false

		if tok.Type == constants.TT_RBRACE && len(blockStack) > 0 {
			closingInline = blockStack[len(blockStack)-1]
		}

		// Whether this token must start a new line. Inline leaf blocks stay cuddled.
		breakLine := false

		if started && prevSig != nil && !commentSincePrevSig && !isComment {
			switch {
			case tok.Type == constants.TT_SEMICOLON:
				breakLine = false
			case tok.Type == constants.TT_LBRACE:
				breakLine = !inline[idx]
			case isElse(tok):
				breakLine = true
			case tok.Type == constants.TT_RBRACE:
				breakLine = !closingInline
			case prevSig.Type == constants.TT_LBRACE:
				breakLine = !parentInline
			case prevSig.Type == constants.TT_RBRACE:
				switch tok.Type {
				case constants.TT_COMMA, constants.TT_RPAREN, constants.TT_RSQUARE,
					constants.TT_LPAREN, constants.TT_LSQUARE:
					breakLine = false
				default:
					breakLine = true
				}
			case prevSig.Type == constants.TT_SEMICOLON && groupDepth == 0 && !parentInline:
				breakLine = true
			}
		}

		// Inline blocks drop their source newlines.
		inlineContext := parentInline || closingInline ||
			(tok.Type == constants.TT_LBRACE && inline[idx])

		if breakLine {
			if pendingNewlines == 0 {
				pendingNewlines = 1
			}
		} else if inlineContext {
			pendingNewlines = 0
		}

		switch {
		case !started:
			pendingNewlines = 0
		case pendingNewlines > 0:
			capN := 2 // at most one blank line

			if tok.Type == constants.TT_RBRACE {
				capN = 1 // none just before a closing brace
			}

			if prevSig != nil && prevSig.Type == constants.TT_LBRACE {
				capN = 1 // none just after an opening brace
			}

			// None in the middle of a statement, e.g. a wrapped condition.
			if prevSig != nil && !isTerminator(prevSig) && !commentSincePrevSig {
				capN = 1
			}

			writeLineStart(tok, capN)
		case isComment:
			b.WriteString(" ")
		case prevSig != nil && wantSpace(prevSig, tok, prevUnaryMinus):
			b.WriteString(" ")
		}

		b.WriteString(text(tok))
		started = true

		switch tok.Type {
		case constants.TT_LBRACE:
			blockStack = append(blockStack, inline[idx])
			braceDepth++
		case constants.TT_LPAREN, constants.TT_LSQUARE:
			groupDepth++
		case constants.TT_RBRACE:
			if len(blockStack) > 0 {
				blockStack = blockStack[:len(blockStack)-1]
			}
		}

		if isComment {
			commentSincePrevSig = true
		} else {
			prevUnaryMinus = tok.Type == constants.TT_MINUS && !isValueEnd(prevSig)
			prevSig = tok
			commentSincePrevSig = false
		}
	}

	result := strings.TrimRight(b.String(), " \t\n")

	if result != "" {
		result += "\n"
	}

	return result
}
