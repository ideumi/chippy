/*
 *
 * The Chippy formatter engine
 *
 */

package formatter

import (
	"chip-go/internal/lexer"
	"chip-go/internal/parser"
)

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
