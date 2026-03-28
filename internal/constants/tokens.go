/*
 *
 * RR2 - internal/constants/tokens.go
 *
 */

package constants

// Types of tokens
const (
	// Literals

	TT_INT        = "__~int~__"
	TT_FLOAT      = "__~flt~__"
	TT_STRING     = "__~str~__"
	TT_IDENTIFIER = "__~idf~__"

	// Keywords

	TT_KEYWORD = "__~kyw~__"

	// Operators

	TT_PLUS  = "__~pls~__"
	TT_MINUS = "__~mns~__"
	TT_MUL   = "__~mul~__"
	TT_DIV   = "__~div~__"
	TT_MOD   = "__~mod~__" // %
	TT_POW   = "__~pow~__" // ^
	TT_EE    = "__~ee~__"  // ==
	TT_NE    = "__~ne~__"  // !=
	TT_LT    = "__~lt~__"  // <
	TT_GT    = "__~gt~__"  // >
	TT_LTE   = "__~lte~__" // <=
	TT_GTE   = "__~gte~__" // >=

	// Assignment

	TT_EQ = "__~eq~__" // =

	// Delimiters

	TT_LPAREN    = "__~lpa~__" // (
	TT_RPAREN    = "__~rpa~__" // )
	TT_LSQUARE   = "__~lsq~__" // [
	TT_RSQUARE   = "__~rsq~__" // ]
	TT_LBRACE    = "__~lbr~__" // {
	TT_RBRACE    = "__~rbr~__" // }
	TT_COMMA     = "__~cma~__" // ,
	TT_COLON     = "__~col~__" // :
	TT_SEMICOLON = "__~smc~__" // ;
	TT_NEWLINE   = "__~nl~__"  // \n

	// Special

	TT_EOF = "__~eof~__" // End of file
)

var Keywords = []string{
	"var",
	"and", "or", "not",
	"if", "elseif", "else",
	"for", "to", "step",
	"while",
	"func",
	"return", "continue", "break",
	"b", "m",
}

func IsKeyword(word string) bool {
	for _, keyword := range Keywords {
		if word == keyword {
			return true
		}
	}

	return false
}
