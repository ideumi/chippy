/*
 *
 * The Chippy formatter engine
 *
 */

package formatter

import (
	"chip-go/internal/constants"
	"chip-go/internal/lexer"
	"chip-go/internal/parser"
	"fmt"
	"strings"
)

type comment struct {
	doc  bool
	text string
}

// Format reformats valid source into the Chippy coding style. It rejects input
// that fails 'chippy check', and refuses any output whose code tokens or comments
// differ from the input after formatting.
func Format(filename, src string) (string, error) {
	lex := lexer.NewLexer(filename, src)
	tokens, err := lex.MakeTokens()

	if err != nil {
		return "", err
	}

	if parseErr := parser.NewParser(tokens).Parse().GetError(); parseErr != nil {
		return "", parseErr
	}

	out := formatTokens(src, tokens)

	if err := verifyEquivalent(filename, tokens, out); err != nil {
		return "", err
	}

	return out, nil
}

func formatTokens(src string, tokens []*lexer.Token) string {
	tokens = reorderCuddles(tokens)

	var b strings.Builder

	braceDepth := 0      // open '{'
	groupDepth := 0      // open '(' and '['
	pendingNewlines := 0 // buffered TT_NEWLINE since the last emitted token
	commentSincePrevSig := false
	prevUnaryMinus := false
	started := false

	var prevSig *lexer.Token // last emitted code token

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

		// A statement wrapping onto the next line (outside any open
		// group) gets one extra level, comment-only lines and line-
		// leading closers don't.
		if !isCommentTok(tok) && !isCloser(tok) && groupDepth == 0 &&
			!isTerminator(prevSig) && !commentSincePrevSig {
			indent++
		}

		if indent < 0 {
			indent = 0
		}

		b.WriteString(strings.Repeat("\t", indent))
	}

	for _, tok := range tokens {
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

		cuddle := false

		if started && pendingNewlines > 0 && !commentSincePrevSig && prevSig != nil {
			if tok.Type == constants.TT_LBRACE && prevSig.Type != constants.TT_SEMICOLON {
				cuddle = true
			} else if isElse(tok) && prevSig.Type == constants.TT_RBRACE {
				cuddle = true
			}
		}

		if started && prevSig != nil && !commentSincePrevSig && !isComment {
			forceBreak := false

			switch {
			case prevSig.Type == constants.TT_LBRACE && tok.Type != constants.TT_RBRACE:
				forceBreak = true
			case tok.Type == constants.TT_RBRACE && prevSig.Type != constants.TT_LBRACE:
				forceBreak = true
			case prevSig.Type == constants.TT_SEMICOLON && groupDepth == 0:
				forceBreak = true
			case prevSig.Type == constants.TT_RBRACE &&
				tok.Type != constants.TT_SEMICOLON && !isElse(tok) && !isCloser(tok):
				forceBreak = true
			}

			if forceBreak {
				cuddle = false

				if pendingNewlines == 0 {
					pendingNewlines = 1
				}
			}
		}

		switch {
		case !started:
			pendingNewlines = 0
		case cuddle:
			b.WriteString(" ")
			pendingNewlines = 0
		case pendingNewlines > 0:
			capN := 2 // at most one blank line

			if tok.Type == constants.TT_RBRACE {
				capN = 1 // none just before a closing brace
			}

			if prevSig != nil && prevSig.Type == constants.TT_LBRACE {
				capN = 1 // none just after an opening brace
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
			braceDepth++
		case constants.TT_LPAREN, constants.TT_LSQUARE:
			groupDepth++
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

// reorderCuddles hops every cuddler ('{', 'else', 'elseif') leftward past any
// comments between it and its anchor, so comments never block brace conversion.
func reorderCuddles(tokens []*lexer.Token) []*lexer.Token {
	result := make([]*lexer.Token, 0, len(tokens))

	for _, tok := range tokens {
		if isCuddler(tok) {
			anchor := -1
			hasComment := false

			for k := len(result) - 1; k >= 0; k-- {
				t := result[k]

				if t.Type == constants.TT_NEWLINE {
					continue
				}

				if isCommentTok(t) {
					hasComment = true
					continue
				}

				anchor = k
				break
			}

			if anchor >= 0 && hasComment && cuddleAnchorOK(tok, result, anchor) {
				// Insert tok right after its anchor and move the
				// intervening comments and newlines to its right.
				result = append(result, nil)
				copy(result[anchor+2:], result[anchor+1:])
				result[anchor+1] = tok

				continue
			}
		}

		result = append(result, tok)
	}

	return result
}

func isCuddler(t *lexer.Token) bool {
	return t.Type == constants.TT_LBRACE || isElse(t)
}

func isElse(t *lexer.Token) bool {
	return t.Type == constants.TT_KEYWORD && (t.Value == "else" || t.Value == "elseif")
}

func isCommentTok(t *lexer.Token) bool {
	return t.Type == constants.TT_COMMENT || t.Type == constants.TT_DOC_COMMENT
}

// cuddleAnchorOK decides whether tok may cuddle onto its anchor. A '{' cuddles
// onto anything but a bare ';'. An else/elseif cuddles only onto a '}' that began
// its own line, which is stricter than the emitter to keep the reorder safe.
func cuddleAnchorOK(tok *lexer.Token, result []*lexer.Token, idx int) bool {
	anchor := result[idx]

	if tok.Type == constants.TT_LBRACE {
		return anchor.Type != constants.TT_SEMICOLON
	}

	if anchor.Type != constants.TT_RBRACE {
		return false
	}

	return idx == 0 || result[idx-1].Type == constants.TT_NEWLINE
}

// isTerminator reports whether t ends a statement, nil counts (start of file).
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

	// '(' / '[' cuddle as a call/index after a value, an m/b literal
	// keyword, or a 'func' parameter list, after any other keyword or operator
	// they space.
	if cur.Type == constants.TT_LPAREN || cur.Type == constants.TT_LSQUARE {
		switch prev.Type {
		case constants.TT_IDENTIFIER, constants.TT_RPAREN, constants.TT_RSQUARE,
			constants.TT_INT, constants.TT_FLOAT, constants.TT_STRING:
			return false
		case constants.TT_KEYWORD:
			if prev.Value == "m" || prev.Value == "b" || prev.Value == "func" {
				return false
			}
		}

		return true
	}

	return true
}

// verifyEquivalent is a safety system: the output must carry the same code tokens
// in order and the same comments. Repositioning a comment is fine, but dropping,
// adding, or altering one is not.
func verifyEquivalent(filename string, orig []*lexer.Token, out string) error {
	lex := lexer.NewLexer(filename, out)
	got, err := lex.MakeTokens()

	if err != nil {
		return bugError("output failed to lex: %w", err)
	}

	want := codeTokens(orig)
	have := codeTokens(got)

	if len(want) != len(have) {
		return bugError("code token count changed (%d -> %d)", len(want), len(have))
	}

	for i := range want {
		if want[i].Type != have[i].Type || want[i].Value != have[i].Value {
			return bugError("code token %d changed (%s -> %s)", i, want[i].String(), have[i].String())
		}
	}

	wantC := commentCounts(orig)
	haveC := commentCounts(got)

	for c, n := range wantC {
		if haveC[c] != n {
			return bugError("a comment was dropped or altered")
		}
	}

	for c, n := range haveC {
		if wantC[c] != n {
			return bugError("a comment was added or altered")
		}
	}

	return nil
}

func bugError(format string, args ...interface{}) error {
	return fmt.Errorf("internal formatter error: "+format+", file was not modified, this is a bug, please report it", args...)
}

func codeTokens(toks []*lexer.Token) []*lexer.Token {
	var out []*lexer.Token

	for _, t := range toks {
		switch t.Type {
		case constants.TT_NEWLINE, constants.TT_EOF, constants.TT_COMMENT, constants.TT_DOC_COMMENT:
			continue
		}

		out = append(out, t)
	}

	return out
}

func commentCounts(toks []*lexer.Token) map[comment]int {
	counts := make(map[comment]int)

	for _, t := range toks {
		switch t.Type {
		case constants.TT_COMMENT:
			counts[comment{false, commentText(t)}]++
		case constants.TT_DOC_COMMENT:
			counts[comment{true, commentText(t)}]++
		}
	}

	return counts
}

func commentText(t *lexer.Token) string {
	return strings.TrimRight(fmt.Sprintf("%v", t.Value), " \t")
}
