/*
 *
 * The Chippy formatter engine / quality control
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

// verifyEquivalent is a safety system: the output must carry the same code tokens
// in order and the same comments, and it must still parse. Repositioning a comment
// is fine, but dropping, adding, or altering one is not.
func verifyEquivalent(filename string, orig []*lexer.Token, out string) error {
	lex := lexer.NewLexer(filename, out)
	got, err := lex.MakeTokens()

	if err != nil {
		return bugError("output failed to lex: %w", err)
	}

	if parseErr := parser.NewParser(got).Parse().GetError(); parseErr != nil {
		return bugError("output failed to parse: %v", parseErr)
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
	s, _ := t.Value.(string)
	return strings.TrimRight(s, " \t")
}
