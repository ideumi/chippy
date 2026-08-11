/*
 *
 * The Chippy formatter engine / inline part
 *
 */

package formatter

import (
	"chip-go/internal/constants"
	"chip-go/internal/lexer"
	"unicode/utf8"
)

// A func body always expands. Any other block collapses when it is a
// short single statement (see leafInlineable) whose one-line form fits the budget.
func inlineBlocks(src string, tokens []*lexer.Token) map[int]bool {
	inline := make(map[int]bool)

	// The first '{' after a 'func' keyword is that function's body.
	funcBody := make(map[int]bool)
	pendingFunc := false

	for i, tok := range tokens {
		switch {
		case tok.Type == constants.TT_KEYWORD && tok.Value == "func":
			pendingFunc = true
		case tok.Type == constants.TT_LBRACE && pendingFunc:
			funcBody[i] = true
			pendingFunc = false
		}
	}

	var stack []int

	for i, tok := range tokens {
		switch tok.Type {
		case constants.TT_LBRACE:
			stack = append(stack, i)
		case constants.TT_RBRACE:
			if len(stack) == 0 {
				continue
			}

			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			indent := len(stack)
			start := headerStart(tokens, open)

			if !funcBody[open] && leafInlineable(tokens, open, i) &&
				!hasComment(tokens, start, open) && !headerWrapped(tokens, start, open) &&
				inlineLineWidth(src, tokens, start, i, indent) <= constants.FORMATTER_INLINE_BUDGET {
				inline[open] = true
			}
		}
	}

	expandChains(tokens, inline)

	return inline
}

// expandChains keeps an if/elseif/else chain uniform: if any branch is not
// inline, none of its branches are. It only ever removes entries from inline.
func expandChains(tokens []*lexer.Token, inline map[int]bool) {
	head := make(map[int]int)       // a branch's '{' index -> its chain's representative
	lastClosed := make(map[int]int) // depth -> '{' index of the last block closed there

	var stack []int

	for i, tok := range tokens {
		switch tok.Type {
		case constants.TT_LBRACE:
			head[i] = i

			if isElse(tokens[headerStart(tokens, i)]) {
				if prev, ok := lastClosed[len(stack)]; ok {
					head[i] = head[prev]
				}
			}

			stack = append(stack, i)
		case constants.TT_RBRACE:
			if len(stack) == 0 {
				continue
			}

			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			lastClosed[len(stack)] = open
		}
	}

	members := make(map[int][]int)

	for open, rep := range head {
		members[rep] = append(members[rep], open)
	}

	for _, group := range members {
		expand := false

		for _, open := range group {
			if !inline[open] {
				expand = true
				break
			}
		}

		if expand {
			for _, open := range group {
				delete(inline, open)
			}
		}
	}
}

// leafInlineable reports whether the block tokens[open+1:closeIdx] fits on one
// line: no nested block, no comment (a '#' line comment would swallow the closing
// brace), at most one statement, and no newline among the body's code tokens.
func leafInlineable(tokens []*lexer.Token, open, closeIdx int) bool {
	semis := 0
	firstBody := -1
	lastBody := -1

	for idx := open + 1; idx < closeIdx; idx++ {
		switch tokens[idx].Type {
		case constants.TT_LBRACE, constants.TT_RBRACE:
			return false
		case constants.TT_COMMENT, constants.TT_DOC_COMMENT:
			return false
		case constants.TT_NEWLINE:
			continue
		case constants.TT_SEMICOLON:
			semis++
		}

		if firstBody == -1 {
			firstBody = idx
		}

		lastBody = idx
	}

	if semis > 1 {
		return false
	}

	// The statement itself must not wrap, so no newline between its first
	// and last code token. An empty block (firstBody == -1) fits.
	for idx := firstBody; firstBody != -1 && idx <= lastBody; idx++ {
		if tokens[idx].Type == constants.TT_NEWLINE {
			return false
		}
	}

	return true
}

// A block can open inside a group (an if-expression used as a value, e.g.
// '(if c {1;} else {2;})[1]'). The walk then exits through the '(' before its
// ')' and depth goes negative, so the boundary test never fires and it runs back
// to the start of the whole statement. That is the right header anyway. Worst
// case a bad guess expands a block that could have stayed inline.
func headerStart(tokens []*lexer.Token, open int) int {
	depth := 0
	first := open

	for idx := open - 1; idx >= 0; idx-- {
		switch tokens[idx].Type {
		case constants.TT_NEWLINE, constants.TT_COMMENT, constants.TT_DOC_COMMENT:
			continue
		case constants.TT_RPAREN, constants.TT_RSQUARE:
			depth++
		case constants.TT_LPAREN, constants.TT_LSQUARE:
			depth--
		}

		if depth == 0 {
			switch tokens[idx].Type {
			case constants.TT_SEMICOLON, constants.TT_LBRACE, constants.TT_RBRACE:
				return first
			}
		}

		first = idx
	}

	return first
}

// A header comment would consume the inlined '{', so a block with one stays
// expanded.
func hasComment(tokens []*lexer.Token, start, end int) bool {
	for idx := start; idx < end; idx++ {
		if isCommentTok(tokens[idx]) {
			return true
		}
	}

	return false
}

// headerWrapped reports whether the header tokens[start:open) span more than one
// line, that is, a newline falls between its code tokens rather than merely before
// the '{'. A wrapped header (such as a multi-line 'if' condition) keeps its block
// expanded, so the result is never half wrapped and half inline.
func headerWrapped(tokens []*lexer.Token, start, open int) bool {
	first := -1
	last := -1

	for idx := start; idx < open; idx++ {
		switch tokens[idx].Type {
		case constants.TT_NEWLINE, constants.TT_COMMENT, constants.TT_DOC_COMMENT:
			continue
		}

		if first == -1 {
			first = idx
		}

		last = idx
	}

	for idx := first; first != -1 && idx <= last; idx++ {
		if tokens[idx].Type == constants.TT_NEWLINE {
			return true
		}
	}

	return false
}

// It applies the same spacing rules as the emitter.
func inlineLineWidth(src string, tokens []*lexer.Token, start, closeIdx, indent int) int {
	width := indent * constants.FORMATTER_TAB_WIDTH

	var prev *lexer.Token
	prevUnaryMinus := false

	for idx := start; idx <= closeIdx; idx++ {
		tok := tokens[idx]

		switch tok.Type {
		case constants.TT_NEWLINE, constants.TT_COMMENT, constants.TT_DOC_COMMENT:
			continue
		}

		if prev != nil && wantSpace(prev, tok, prevUnaryMinus) {
			width++
		}

		width += utf8.RuneCountInString(src[tok.PosStart.Index:tok.PosEnd.Index])
		prevUnaryMinus = tok.Type == constants.TT_MINUS && !isValueEnd(prev)
		prev = tok
	}

	return width
}
