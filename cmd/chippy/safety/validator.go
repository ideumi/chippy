/*
 *
 * Combine's validator for common issues and
 * easy to make and hard to spot mistakes (C.V.C.I.A.E.M.H.S.M, for reference)
 *
 */

package safety

import (
	"chip-go/internal/ast"
	"chip-go/internal/builtins"
	"chip-go/internal/lexer"
	"chip-go/internal/parser"
	"fmt"
	"os"
)

type SymbolInfo struct {
	Name    string
	File    string
	Line    int
	SymType string // "function" or "variable"
}

type ValidationResult struct {
	Collisions []SymbolCollision
	Errors     []ValidationError
}

type SymbolCollision struct {
	Name      string
	SymType   string
	Locations []SymbolInfo
}

func (c *SymbolCollision) IsSameFile() bool {
	if len(c.Locations) == 0 {
		return false
	}

	firstFile := c.Locations[0].File

	for _, loc := range c.Locations[1:] {
		if loc.File != firstFile {
			return false
		}
	}

	return true
}

func (c *SymbolCollision) GetFirstFile() string {
	if len(c.Locations) > 0 {
		return c.Locations[0].File
	}

	return ""
}

type ValidationError struct {
	File  string
	Error error
}

var builtinNames = initBuiltinNames()

func initBuiltinNames() map[string]string {
	names := make(map[string]string)

	for name := range builtins.GetBuiltins() {
		names[name] = "builtin function"
	}

	for name := range builtins.GetConstants() {
		names[name] = "builtin constant"
	}

	return names
}

func (c *SymbolCollision) IsBuiltinCollision() bool {
	return c.SymType == "builtin function" || c.SymType == "builtin constant"
}

func ValidateFiles(files []string) (*ValidationResult, error) {
	result := &ValidationResult{
		Collisions: []SymbolCollision{},
		Errors:     []ValidationError{},
	}

	symbolRegistry := make(map[string][]SymbolInfo)

	for _, file := range files {
		symbols, err := extractSymbols(file)

		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				File:  file,
				Error: err,
			})

			continue
		}

		for _, sym := range symbols {
			symbolRegistry[sym.Name] = append(symbolRegistry[sym.Name], sym)
		}
	}

	for name, locations := range symbolRegistry {
		if builtinType, isBuiltin := builtinNames[name]; isBuiltin {
			result.Collisions = append(result.Collisions, SymbolCollision{
				Name:      name,
				SymType:   builtinType,
				Locations: locations,
			})
			continue
		}

		if len(locations) > 1 {
			symType := locations[0].SymType

			for _, loc := range locations[1:] {
				if loc.SymType != symType {
					symType = "symbol"

					break
				}
			}

			result.Collisions = append(result.Collisions, SymbolCollision{
				Name:      name,
				SymType:   symType,
				Locations: locations,
			})
		}
	}

	return result, nil
}

func extractSymbols(filename string) ([]SymbolInfo, error) {
	content, err := os.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	lex := lexer.NewLexer(filename, string(content))
	tokens, err := lex.MakeTokens()

	if err != nil {
		return nil, err
	}

	parseResult := parser.NewParser(tokens).Parse()

	if parseResult.GetError() != nil {
		return nil, parseResult.GetError()
	}

	node := parseResult.GetNode()
	symbols := []SymbolInfo{}

	walkAST(node, filename, &symbols)

	return symbols, nil
}

// Walk the AST and extract file scope symbol idfs. See 'chippy doc scoping' to
// understand the scoping desicions here better
func walkAST(node ast.Node, filename string, symbols *[]SymbolInfo) {
	if node == nil {
		return
	}

	switch typed := node.(type) {
	case *ast.FuncDefNode:
		if typed.VarNameToken != nil {
			name, ok := typed.VarNameToken.Value.(string)

			if !ok {
				return
			}

			*symbols = append(*symbols, SymbolInfo{
				Name:    name,
				File:    filename,
				Line:    typed.VarNameToken.PosStart.DisplayLine(),
				SymType: "function",
			})
		}
		// Function body has isolated scope, stop

	case *ast.VarAssignNode:
		name, ok := typed.VarNameToken.Value.(string)

		if !ok {
			return
		}

		*symbols = append(*symbols, SymbolInfo{
			Name:    name,
			File:    filename,
			Line:    typed.VarNameToken.PosStart.DisplayLine(),
			SymType: "variable",
		})

	// Jump in since we are still in file scope:
	case *ast.IfNode:
		for _, caseClause := range typed.Cases {
			walkAST(caseClause.Body, filename, symbols)
		}

		if typed.ElseCase != nil {
			walkAST(typed.ElseCase, filename, symbols)
		}

	case *ast.ForNode:
		walkAST(typed.BodyNode, filename, symbols)

	case *ast.WhileNode:
		walkAST(typed.BodyNode, filename, symbols)

	case *ast.ListNode:
		for _, elem := range typed.ElementNodes {
			walkAST(elem, filename, symbols)
		}

	case *ast.BlockNode:
		// Block bodies and the top-level program are file scope, keep
		// walking
		for _, elem := range typed.ElementNodes {
			walkAST(elem, filename, symbols)
		}
	}
}
