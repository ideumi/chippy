/*
 *
 * Modena - internal/compiler/compile.go
 *
 */

package compiler

import (
	"chip-go/internal/bytecode"
	"chip-go/internal/bytecode/serializer"
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

	return serializer.Serialize(chunk, addShebang), nil
}

func DecodeOrCompile(filename string, data []byte) (*bytecode.Chunk, error) {
	if serializer.LooksLikeBytecode(data) {
		return serializer.Deserialize(data)
	}

	return CompileSource(filename, string(data))
}
