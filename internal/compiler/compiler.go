/*
 *
 * Modena - internal/compiler/compiler.go
 *
 */

package compiler

import (
	"chip-go/internal/bytecode"
	"chip-go/internal/lexer"
	"chip-go/internal/parser"
)

func CompileSource(filename, text string) (*bytecode.Chunk, error) {
	lex := lexer.NewLexer(filename, text)
	tokens, err := lex.MakeTokens()

	if err != nil {
		return nil, err
	}

	parseResult := parser.NewParser(tokens).Parse()

	if parseResult.GetError() != nil {
		return nil, parseResult.GetError()
	}

	return NewCompiler().Compile(parseResult.GetNode())
}

func CompileToBytecode(filename, text string, addShebang bool) ([]byte, error) {
	chunk, err := CompileSource(filename, text)

	if err != nil {
		return nil, err
	}

	return bytecode.Serialize(chunk, addShebang), nil
}

func DecodeOrCompile(filename string, data []byte) (*bytecode.Chunk, error) {
	if bytecode.LooksLikeBytecode(data) {
		return bytecode.Deserialize(data)
	}

	return CompileSource(filename, string(data))
}
