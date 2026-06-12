/*
 *
 * RR2 - internal/parser/parser.go
 *
 */

// Behold, the kraken!

package parser

import (
	"chip-go/internal/ast"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/lexer"
)

type ifExprResult struct {
	*ast.BaseNode
	cases    []ast.IfCase
	elseCase ast.Node
}

func (i *ifExprResult) String() string {
	return "ifExprResult"
}

type Parser struct {
	tokens     []*lexer.Token
	tokIdx     int
	currentTok *lexer.Token
}

func NewParser(tokens []*lexer.Token) *Parser {
	parser := &Parser{
		tokens: tokens,
		tokIdx: -1,
	}
	parser.advance()

	return parser
}

func (p *Parser) advance() *lexer.Token {
	p.tokIdx++
	p.updateCurrentTok()

	return p.currentTok
}

func (p *Parser) reverse(amount int) *lexer.Token {
	p.tokIdx -= amount
	p.updateCurrentTok()

	return p.currentTok
}

func (p *Parser) updateCurrentTok() {
	if p.tokIdx >= 0 && p.tokIdx < len(p.tokens) {
		p.currentTok = p.tokens[p.tokIdx]
	} else {
		p.currentTok = nil
	}
}

func isWhitespace(tok *lexer.Token) bool {
	return tok != nil && (tok.Type == constants.TT_NEWLINE || tok.Type == constants.TT_COMMENT || tok.Type == constants.TT_DOC_COMMENT)
}

func (p *Parser) skipWhitespace(res *ParseResult) {
	for isWhitespace(p.currentTok) {
		res.RegisterAdvancement()
		p.advance()
	}
}

func (p *Parser) peekPastWhitespace() *lexer.Token {
	for i := p.tokIdx; i >= 0 && i < len(p.tokens); i++ {
		if !isWhitespace(p.tokens[i]) {
			return p.tokens[i]
		}
	}

	return nil
}

func (p *Parser) Parse() *ParseResult {
	res := p.statements()
	if res.error != nil {
		return res
	}

	p.skipWhitespace(res)

	// Only error if there are non-EOF tokens remaining that aren't newlines or comments
	if p.currentTok != nil && p.currentTok.Type != constants.TT_EOF {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Unexpected token",
		))
	}

	return res
}

func (p *Parser) statements() *ParseResult {
	res := NewParseResult()
	statements := []ast.Node{}

	if p.currentTok == nil {
		return res.Failure(errors.NewInvalidSyntaxError(
			nil, nil,
			"Unexpected end of input",
		))
	}

	posStart := p.currentTok.PosStart.Copy()

	p.skipWhitespace(res)

	// Parse statements until we hit a closing brace or EOF
	for p.currentTok != nil && p.currentTok.Type != constants.TT_RBRACE && p.currentTok.Type != constants.TT_EOF {
		// Parse a single statement
		statement := res.Register(p.statement())

		if res.error != nil {
			return res
		}

		statements = append(statements, statement)

		// Check if this is a compound statement, they don't require semicolons after their closing brace
		isCompoundStatement := false

		switch statement.(type) {
		case *ast.IfNode, *ast.ForNode, *ast.WhileNode, *ast.FuncDefNode:
			isCompoundStatement = true
		}

		if !isCompoundStatement {
			// Enforce semicolons
			if p.currentTok == nil || p.currentTok.Type != constants.TT_SEMICOLON {
				return res.Failure(errors.NewInvalidSyntaxError(
					p.currentTok.PosStart, p.currentTok.PosEnd,
					"Expected ';' after statement",
				))
			}

			res.RegisterAdvancement()
			p.advance()
		} else {
			// Compound statements don't require semicolons, but allow them optionally
			if p.currentTok != nil && p.currentTok.Type == constants.TT_SEMICOLON {
				res.RegisterAdvancement()
				p.advance()
			}
		}

		p.skipWhitespace(res)
	}

	if len(statements) == 0 {
		return res.Success(ast.NewListNode(statements, posStart, posStart))
	}

	return res.Success(ast.NewListNode(statements, posStart, statements[len(statements)-1].GetPosEnd()))
}

func (p *Parser) statement() *ParseResult {
	res := NewParseResult()
	posStart := p.currentTok.PosStart.Copy()

	if p.currentTok.Matches(constants.TT_KEYWORD, "return") {
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		expr := res.TryRegister(p.expr())

		if expr == nil {
			p.reverse(res.GetToReverseCount())
		}

		return res.Success(ast.NewReturnNode(expr, posStart, p.currentTok.PosStart.Copy()))
	}

	if p.currentTok.Matches(constants.TT_KEYWORD, "continue") {
		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewContinueNode(posStart, p.currentTok.PosStart.Copy()))
	}

	if p.currentTok.Matches(constants.TT_KEYWORD, "break") {
		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewBreakNode(posStart, p.currentTok.PosStart.Copy()))
	}

	if p.currentTok.Matches(constants.TT_KEYWORD, "var") {
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		if p.currentTok.Type != constants.TT_IDENTIFIER {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected identifier",
			))
		}

		varName := p.currentTok
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		if p.currentTok.Type != constants.TT_EQ {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected '='",
			))
		}

		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		valueExpr := res.Register(p.expr())

		if res.error != nil {
			return res
		}

		return res.Success(ast.NewVarAssignNode(varName, valueExpr))
	}

	node := res.Register(p.expr())
	if res.error != nil {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected 'return', 'continue', 'break', 'var', 'if', 'for', 'while', 'func', int, float, identifier, '+', '-', '(', '[' or 'not'",
		))
	}

	p.skipWhitespace(res)

	if p.currentTok != nil && p.currentTok.Type == constants.TT_EQ {
		switch target := node.(type) {
		case *ast.VarAccessNode:
			res.RegisterAdvancement()
			p.advance()

			p.skipWhitespace(res)

			valueExpr := res.Register(p.expr())

			if res.error != nil {
				return res
			}

			return res.Success(ast.NewVarUpdateNode(target.VarNameToken, valueExpr))

		case *ast.IndexAccessNode:
			res.RegisterAdvancement()
			p.advance()

			p.skipWhitespace(res)

			valueExpr := res.Register(p.expr())

			if res.error != nil {
				return res
			}

			return res.Success(ast.NewIndexAssignNode(target.CollectionNode, target.IndexNode, valueExpr))
		}
	}

	return res.Success(node)
}

func (p *Parser) expr() *ParseResult {
	res := NewParseResult()

	node := res.Register(p.binOp(p.bitwiseExpr, []string{constants.TT_KEYWORD}, []interface{}{"and", "or", "xor"}))

	if res.error != nil {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected 'var', 'if', 'for', 'while', 'func', int, float, identifier, '+', '-', '(', '[', 'not' or 'bnot'",
		))
	}

	return res.Success(node)
}

func (p *Parser) bitwiseExpr() *ParseResult {
	return p.binOp(p.compExpr, []string{constants.TT_KEYWORD}, []interface{}{"band", "bor", "bxor"})
}

func (p *Parser) compExpr() *ParseResult {
	res := NewParseResult()

	if p.currentTok.Matches(constants.TT_KEYWORD, "not") {
		opTok := p.currentTok
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		node := res.Register(p.compExpr())
		if res.error != nil {
			return res
		}

		return res.Success(ast.NewUnaryOpNode(opTok, node))
	}

	node := res.Register(p.binOp(p.shiftExpr, []string{constants.TT_EE, constants.TT_NE, constants.TT_LT, constants.TT_GT, constants.TT_LTE, constants.TT_GTE}, nil))

	if res.error != nil {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected int, float, identifier, '+', '-', '(', '[', 'if', 'for', 'while', 'func', 'not' or 'bnot'",
		))
	}

	return res.Success(node)
}

func (p *Parser) shiftExpr() *ParseResult {
	return p.binOp(p.arithExpr, []string{constants.TT_LSHIFT, constants.TT_RSHIFT}, nil)
}

func (p *Parser) arithExpr() *ParseResult {
	return p.binOp(p.term, []string{constants.TT_PLUS, constants.TT_MINUS}, nil)
}

func (p *Parser) term() *ParseResult {
	return p.binOp(p.factor, []string{constants.TT_MUL, constants.TT_DIV, constants.TT_MOD}, nil)
}

func (p *Parser) factor() *ParseResult {
	res := NewParseResult()
	tok := p.currentTok

	if tok.Type == constants.TT_PLUS || tok.Type == constants.TT_MINUS || tok.Matches(constants.TT_KEYWORD, "bnot") {
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		factor := res.Register(p.factor())

		if res.error != nil {
			return res
		}

		return res.Success(ast.NewUnaryOpNode(tok, factor))
	}

	return p.power()
}

func (p *Parser) power() *ParseResult {
	return p.binOp(p.call, []string{constants.TT_POW}, nil)
}

func (p *Parser) call() *ParseResult {
	res := NewParseResult()
	atom := res.Register(p.atom())

	if res.error != nil {
		return res
	}

	// Handle function calls with () and indexing with []
	for {
		// Handle function calls with ()
		if p.currentTok.Type == constants.TT_LPAREN {
			res.RegisterAdvancement()
			p.advance()
			argNodes := []ast.Node{}

			p.skipWhitespace(res)

			if p.currentTok.Type == constants.TT_RPAREN {
				res.RegisterAdvancement()
				p.advance()
			} else {
				argNodes = append(argNodes, res.Register(p.expr()))

				if res.error != nil {
					return res.Failure(errors.NewInvalidSyntaxError(
						p.currentTok.PosStart, p.currentTok.PosEnd,
						"Expected ')', 'var', 'if', 'for', 'while', 'func', int, float, identifier, '+', '-', '(', '[' or 'not'",
					))
				}

				for p.currentTok.Type == constants.TT_COMMA {
					res.RegisterAdvancement()
					p.advance()

					p.skipWhitespace(res)

					// Allow trailing comma (closing paren after comma)
					if p.currentTok.Type == constants.TT_RPAREN {
						break
					}

					argNodes = append(argNodes, res.Register(p.expr()))
					if res.error != nil {
						return res
					}
				}

				p.skipWhitespace(res)

				if p.currentTok.Type != constants.TT_RPAREN {
					return res.Failure(errors.NewInvalidSyntaxError(
						p.currentTok.PosStart, p.currentTok.PosEnd,
						"Expected ',' or ')'",
					))
				}

				res.RegisterAdvancement()
				p.advance()
			}

			atom = ast.NewCallNode(atom, argNodes)

			continue
		}

		// Handle indexing with []
		if p.currentTok.Type == constants.TT_LSQUARE {
			res.RegisterAdvancement()
			p.advance()

			p.skipWhitespace(res)

			indexNode := res.Register(p.expr())
			if res.error != nil {
				return res
			}

			p.skipWhitespace(res)

			if p.currentTok.Type != constants.TT_RSQUARE {
				return res.Failure(errors.NewInvalidSyntaxError(
					p.currentTok.PosStart, p.currentTok.PosEnd,
					"Expected ']'",
				))
			}

			posEnd := p.currentTok.PosEnd
			res.RegisterAdvancement()
			p.advance()

			atom = ast.NewIndexAccessNode(atom, indexNode, posEnd)

			continue
		}

		break
	}

	return res.Success(atom)
}

func (p *Parser) atom() *ParseResult {
	res := NewParseResult()
	tok := p.currentTok

	if tok.Type == constants.TT_INT || tok.Type == constants.TT_FLOAT {
		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewNumberNode(tok))
	}

	if tok.Type == constants.TT_STRING {
		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewStringNode(tok))
	}

	if tok.Type == constants.TT_IDENTIFIER {
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		return res.Success(ast.NewVarAccessNode(tok))
	}

	if tok.Type == constants.TT_LPAREN {
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		expr := res.Register(p.expr())

		if res.error != nil {
			return res
		}

		p.skipWhitespace(res)

		if p.currentTok.Type == constants.TT_RPAREN {
			res.RegisterAdvancement()
			p.advance()

			return res.Success(expr)
		} else {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected ')'",
			))
		}
	}

	if tok.Type == constants.TT_LSQUARE {
		listExpr := res.Register(p.listExpr())

		if res.error != nil {
			return res
		}

		return res.Success(listExpr)
	}

	if tok.Matches(constants.TT_KEYWORD, "b") {
		res.RegisterAdvancement()
		p.advance()
		byteArrayExpr := res.Register(p.byteArrayExpr())

		if res.error != nil {
			return res
		}

		return res.Success(byteArrayExpr)
	}

	if tok.Matches(constants.TT_KEYWORD, "m") {
		res.RegisterAdvancement()
		p.advance()
		mapExpr := res.Register(p.mapExpr())

		if res.error != nil {
			return res
		}

		return res.Success(mapExpr)
	}

	if tok.Matches(constants.TT_KEYWORD, "if") {
		ifExpr := res.Register(p.ifExpr())

		if res.error != nil {
			return res
		}

		return res.Success(ifExpr)
	}

	if tok.Matches(constants.TT_KEYWORD, "for") {
		forExpr := res.Register(p.forExpr())

		if res.error != nil {
			return res
		}

		return res.Success(forExpr)
	}

	if tok.Matches(constants.TT_KEYWORD, "while") {
		whileExpr := res.Register(p.whileExpr())

		if res.error != nil {
			return res
		}

		return res.Success(whileExpr)
	}

	if tok.Matches(constants.TT_KEYWORD, "func") {
		funcDef := res.Register(p.funcDef())

		if res.error != nil {
			return res
		}

		return res.Success(funcDef)
	}

	return res.Failure(errors.NewInvalidSyntaxError(
		tok.PosStart, tok.PosEnd,
		"Expected int, float, identifier, '+', '-', '(', '[', 'b', 'm', 'if', 'for', 'while', 'func'",
	))
}

func (p *Parser) listExpr() *ParseResult {
	res := NewParseResult()
	elementNodes := []ast.Node{}
	posStart := p.currentTok.PosStart.Copy()

	if p.currentTok.Type != constants.TT_LSQUARE {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '['",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	if p.currentTok.Type == constants.TT_RSQUARE {
		res.RegisterAdvancement()
		p.advance()
	} else {
		elementNodes = append(elementNodes, res.Register(p.expr()))

		if res.error != nil {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected ']', 'var', 'if', 'for', 'while', 'func', int, float, identifier, '+', '-', '(', '[' or 'not'",
			))
		}

		for p.currentTok.Type == constants.TT_COMMA {
			res.RegisterAdvancement()
			p.advance()

			p.skipWhitespace(res)

			// Allow trailing comma (closing bracket after comma)
			if p.currentTok.Type == constants.TT_RSQUARE {
				break
			}

			elementNodes = append(elementNodes, res.Register(p.expr()))
			if res.error != nil {
				return res
			}
		}

		p.skipWhitespace(res)

		if p.currentTok.Type != constants.TT_RSQUARE {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected ',' or ']'",
			))
		}

		// Store the closing bracket position before advancing
		closingBracketEnd := p.currentTok.PosEnd.Copy()
		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewListNode(elementNodes, posStart, closingBracketEnd))
	}

	// For empty lists, use the closing bracket position
	closingBracketEnd := p.currentTok.PosEnd.Copy()

	return res.Success(ast.NewListNode(elementNodes, posStart, closingBracketEnd))
}

func (p *Parser) mapExpr() *ParseResult {
	res := NewParseResult()
	posStart := p.currentTok.PosStart.Copy()

	if p.currentTok.Type != constants.TT_LSQUARE {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '['",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	// Empty map: m[]
	if p.currentTok == nil || p.currentTok.Type == constants.TT_RSQUARE {
		if p.currentTok == nil {
			return res.Failure(errors.NewInvalidSyntaxError(
				posStart, posStart,
				"Expected ']'",
			))
		}

		closingBracketEnd := p.currentTok.PosEnd.Copy()
		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewMapNode(posStart, closingBracketEnd, nil, nil))
	}

	firstExpr := res.Register(p.expr())

	if res.error != nil {
		return res
	}

	p.skipWhitespace(res)

	if p.currentTok != nil && p.currentTok.Type == constants.TT_COLON {
		keyNodes := []ast.Node{firstExpr}
		var valueNodes []ast.Node

		for {
			res.RegisterAdvancement()
			p.advance()

			p.skipWhitespace(res)

			valNode := res.Register(p.expr())

			if res.error != nil {
				return res
			}

			valueNodes = append(valueNodes, valNode)

			p.skipWhitespace(res)

			if p.currentTok == nil || p.currentTok.Type == constants.TT_RSQUARE {
				break
			}

			if p.currentTok.Type != constants.TT_COMMA {
				return res.Failure(errors.NewInvalidSyntaxError(
					p.currentTok.PosStart, p.currentTok.PosEnd,
					"Expected ',' or ']' in map literal",
				))
			}

			res.RegisterAdvancement()
			p.advance()

			p.skipWhitespace(res)

			// Allow trailing comma
			if p.currentTok == nil || p.currentTok.Type == constants.TT_RSQUARE {
				break
			}

			keyNode := res.Register(p.expr())

			if res.error != nil {
				return res
			}

			keyNodes = append(keyNodes, keyNode)

			p.skipWhitespace(res)

			if p.currentTok == nil || p.currentTok.Type != constants.TT_COLON {
				if p.currentTok != nil {
					return res.Failure(errors.NewInvalidSyntaxError(
						p.currentTok.PosStart, p.currentTok.PosEnd,
						"Expected ':' after map key",
					))
				}

				return res.Failure(errors.NewInvalidSyntaxError(
					posStart, posStart,
					"Expected ':' after map key",
				))
			}
		}

		// Store closing bracket position before advancing
		if p.currentTok == nil || p.currentTok.Type != constants.TT_RSQUARE {
			if p.currentTok != nil {
				return res.Failure(errors.NewInvalidSyntaxError(
					p.currentTok.PosStart, p.currentTok.PosEnd,
					"Expected ']'",
				))
			}

			return res.Failure(errors.NewInvalidSyntaxError(
				posStart, posStart,
				"Expected ']'",
			))
		}

		closingBracketEnd := p.currentTok.PosEnd.Copy()
		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewMapNode(posStart, closingBracketEnd, keyNodes, valueNodes))
	}

	if p.currentTok != nil {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected ':' after map key",
		))
	}

	return res.Failure(errors.NewInvalidSyntaxError(
		posStart, posStart,
		"Expected ':' after map key",
	))
}

func (p *Parser) byteArrayExpr() *ParseResult {
	res := NewParseResult()
	elementNodes := []ast.Node{}
	posStart := p.currentTok.PosStart.Copy()

	if p.currentTok.Type != constants.TT_LSQUARE {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '['",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	if p.currentTok.Type == constants.TT_RSQUARE {
		res.RegisterAdvancement()
		p.advance()
	} else {
		elementNodes = append(elementNodes, res.Register(p.expr()))
		if res.error != nil {
			return res
		}

		for p.currentTok.Type == constants.TT_COMMA {
			res.RegisterAdvancement()
			p.advance()

			p.skipWhitespace(res)

			// Allow trailing comma (closing bracket after comma)
			if p.currentTok.Type == constants.TT_RSQUARE {
				break
			}

			elementNodes = append(elementNodes, res.Register(p.expr()))

			if res.error != nil {
				return res
			}
		}

		p.skipWhitespace(res)

		if p.currentTok.Type != constants.TT_RSQUARE {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected ',' or ']'",
			))
		}

		// Store the closing bracket position before advancing
		closingBracketEnd := p.currentTok.PosEnd.Copy()
		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewByteArrayNode(elementNodes, posStart, closingBracketEnd))
	}

	// For empty byte arrays, use the closing bracket position
	closingBracketEnd := p.currentTok.PosEnd.Copy()

	return res.Success(ast.NewByteArrayNode(elementNodes, posStart, closingBracketEnd))
}

func (p *Parser) ifExpr() *ParseResult {
	res := NewParseResult()
	result := p.ifExprCases("if")

	if result.error != nil {
		return result
	}

	ifResult := result.node.(*ifExprResult)

	return res.Success(ast.NewIfNode(ifResult.cases, ifResult.elseCase))
}

func (p *Parser) ifExprCases(caseKeyword string) *ParseResult {
	res := NewParseResult()
	cases := []ast.IfCase{}
	var elseCase ast.Node

	if !p.currentTok.Matches(constants.TT_KEYWORD, caseKeyword) {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '"+caseKeyword+"'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	condition := res.Register(p.expr())
	if res.error != nil {
		return res
	}

	if err := p.expectLBraceWithOptionalNewline(res); err != nil {
		return res.Failure(err)
	}

	if isWhitespace(p.currentTok) {
		res.RegisterAdvancement()
		p.advance()

		statements := res.Register(p.statements())

		if res.error != nil {
			return res
		}

		cases = append(cases, ast.IfCase{
			Condition:        condition,
			Body:             statements,
			ShouldReturnNull: true,
		})

		if p.currentTok.Type != constants.TT_RBRACE {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected '}'",
			))
		}

		res.RegisterAdvancement()
		p.advance()

		// Check for elseif/else continuation
		result := p.tryParseElseifElse()

		if result != nil {
			if result.error != nil {
				return result
			}

			ifResult := result.node.(*ifExprResult)
			cases = append(cases, ifResult.cases...)
			elseCase = ifResult.elseCase
		}
	} else {
		statements := res.Register(p.statements())
		if res.error != nil {
			return res
		}

		cases = append(cases, ast.IfCase{
			Condition:        condition,
			Body:             statements,
			ShouldReturnNull: false,
		})

		if p.currentTok.Type != constants.TT_RBRACE {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected '}'",
			))
		}

		res.RegisterAdvancement()
		p.advance()

		// Check for elseif/else continuation
		result := p.tryParseElseifElse()
		if result != nil {
			if result.error != nil {
				return result
			}

			ifResult := result.node.(*ifExprResult)
			cases = append(cases, ifResult.cases...)
			elseCase = ifResult.elseCase
		}
	}

	return res.Success(&ifExprResult{
		BaseNode: ast.NewBaseNode(nil, nil),
		cases:    cases,
		elseCase: elseCase,
	})
}

func (p *Parser) tryParseElseifElse() *ParseResult {
	// Peek past whitespace so a non-continuation leaves the token stream untouched.
	next := p.peekPastWhitespace()
	if next == nil {
		return nil
	}

	res := NewParseResult()

	if next.Matches(constants.TT_KEYWORD, "elseif") {
		p.skipWhitespace(res)

		return p.ifExprCases("elseif")
	}

	if next.Matches(constants.TT_KEYWORD, "else") {
		p.skipWhitespace(res)

		elseCase := res.Register(p.ifExprC())

		if res.error != nil {
			return res
		}

		return res.Success(&ifExprResult{
			BaseNode: ast.NewBaseNode(nil, nil),
			cases:    nil,
			elseCase: elseCase,
		})
	}

	return nil
}

func (p *Parser) ifExprC() *ParseResult {
	res := NewParseResult()
	var elseCase ast.Node

	if p.currentTok.Matches(constants.TT_KEYWORD, "else") {
		res.RegisterAdvancement()
		p.advance()

		if err := p.expectLBraceWithOptionalNewline(res); err != nil {
			return res.Failure(err)
		}

		if isWhitespace(p.currentTok) {
			res.RegisterAdvancement()
			p.advance()

			statements := res.Register(p.statements())

			if res.error != nil {
				return res
			}

			elseCase = statements

			if p.currentTok.Type != constants.TT_RBRACE {
				return res.Failure(errors.NewInvalidSyntaxError(
					p.currentTok.PosStart, p.currentTok.PosEnd,
					"Expected '}'",
				))
			}

			res.RegisterAdvancement()
			p.advance()
		} else {
			statements := res.Register(p.statements())

			if res.error != nil {
				return res
			}

			elseCase = statements

			if p.currentTok.Type != constants.TT_RBRACE {
				return res.Failure(errors.NewInvalidSyntaxError(
					p.currentTok.PosStart, p.currentTok.PosEnd,
					"Expected '}'",
				))
			}

			res.RegisterAdvancement()
			p.advance()
		}
	}

	return res.Success(elseCase)
}

func (p *Parser) forExpr() *ParseResult {
	res := NewParseResult()

	if !p.currentTok.Matches(constants.TT_KEYWORD, "for") {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected 'for'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	if p.currentTok.Type != constants.TT_IDENTIFIER {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected identifier",
		))
	}

	varName := p.currentTok
	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	if p.currentTok.Type != constants.TT_EQ {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '='",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	startValue := res.Register(p.expr())
	if res.error != nil {
		return res
	}

	if !p.currentTok.Matches(constants.TT_KEYWORD, "to") {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected 'to'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	endValue := res.Register(p.expr())
	if res.error != nil {
		return res
	}

	var stepValue ast.Node
	if p.currentTok.Matches(constants.TT_KEYWORD, "step") {
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		stepValue = res.Register(p.expr())

		if res.error != nil {
			return res
		}
	}

	if err := p.expectLBraceWithOptionalNewline(res); err != nil {
		return res.Failure(err)
	}

	if isWhitespace(p.currentTok) {

		res.RegisterAdvancement()
		p.advance()

		body := res.Register(p.statements())

		if res.error != nil {
			return res
		}

		if p.currentTok.Type != constants.TT_RBRACE {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected '}'",
			))
		}

		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewForNode(varName, startValue, endValue, stepValue, body, true))
	}

	body := res.Register(p.statements())
	if res.error != nil {
		return res
	}

	if p.currentTok.Type != constants.TT_RBRACE {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '}'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	return res.Success(ast.NewForNode(varName, startValue, endValue, stepValue, body, false))
}

func (p *Parser) whileExpr() *ParseResult {
	res := NewParseResult()

	if !p.currentTok.Matches(constants.TT_KEYWORD, "while") {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected 'while'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	condition := res.Register(p.expr())
	if res.error != nil {
		return res
	}

	if err := p.expectLBraceWithOptionalNewline(res); err != nil {
		return res.Failure(err)
	}

	if isWhitespace(p.currentTok) {
		res.RegisterAdvancement()
		p.advance()

		body := res.Register(p.statements())

		if res.error != nil {
			return res
		}

		if p.currentTok.Type != constants.TT_RBRACE {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected '}'",
			))
		}

		res.RegisterAdvancement()
		p.advance()

		return res.Success(ast.NewWhileNode(condition, body, true))
	}

	body := res.Register(p.statements())
	if res.error != nil {
		return res
	}

	if p.currentTok.Type != constants.TT_RBRACE {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '}'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	return res.Success(ast.NewWhileNode(condition, body, false))
}

func (p *Parser) funcDef() *ParseResult {
	res := NewParseResult()

	if !p.currentTok.Matches(constants.TT_KEYWORD, "func") {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected 'func'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	var varNameTok *lexer.Token
	if p.currentTok.Type == constants.TT_IDENTIFIER {
		varNameTok = p.currentTok

		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		if p.currentTok.Type != constants.TT_LPAREN {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected '('",
			))
		}
	} else {
		if p.currentTok.Type != constants.TT_LPAREN {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected identifier or '('",
			))
		}
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	argNameToks := []*lexer.Token{}

	if p.currentTok.Type == constants.TT_IDENTIFIER {
		argNameToks = append(argNameToks, p.currentTok)

		res.RegisterAdvancement()
		p.advance()

		for p.currentTok.Type == constants.TT_COMMA {
			res.RegisterAdvancement()
			p.advance()

			p.skipWhitespace(res)

			// Allow trailing comma (closing paren after comma)
			if p.currentTok.Type == constants.TT_RPAREN {
				break
			}

			if p.currentTok.Type != constants.TT_IDENTIFIER {
				return res.Failure(errors.NewInvalidSyntaxError(
					p.currentTok.PosStart, p.currentTok.PosEnd,
					"Expected identifier",
				))
			}

			argNameToks = append(argNameToks, p.currentTok)
			res.RegisterAdvancement()
			p.advance()
		}

		p.skipWhitespace(res)

		if p.currentTok.Type != constants.TT_RPAREN {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected ',' or ')'",
			))
		}
	} else {
		if p.currentTok.Type != constants.TT_RPAREN {
			return res.Failure(errors.NewInvalidSyntaxError(
				p.currentTok.PosStart, p.currentTok.PosEnd,
				"Expected identifier or ')'",
			))
		}
	}

	res.RegisterAdvancement()
	p.advance()

	p.skipWhitespace(res)

	if p.currentTok.Type != constants.TT_LBRACE {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '{'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	body := res.Register(p.statements())
	if res.error != nil {
		return res
	}

	if p.currentTok.Type != constants.TT_RBRACE {
		return res.Failure(errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '}'",
		))
	}

	res.RegisterAdvancement()
	p.advance()

	return res.Success(ast.NewFuncDefNode(varNameTok, argNameToks, body, false))
}

func (p *Parser) binOp(leftFunc func() *ParseResult, ops []string, opValues []interface{}) *ParseResult {
	res := NewParseResult()
	left := res.Register(leftFunc())
	if res.error != nil {
		return res
	}

	p.skipWhitespace(res)

	for p.containsOp(ops, opValues) {
		opTok := p.currentTok
		res.RegisterAdvancement()
		p.advance()

		p.skipWhitespace(res)

		right := res.Register(leftFunc())

		if res.error != nil {
			return res
		}

		left = ast.NewBinOpNode(left, opTok, right)

		p.skipWhitespace(res)
	}

	return res.Success(left)
}

func (p *Parser) containsOp(ops []string, opValues []interface{}) bool {
	for _, op := range ops {
		if opValues != nil {
			// Check if current token matches this operator type with any of the values
			for _, value := range opValues {
				if p.currentTok.Matches(op, value) {
					return true
				}
			}
		} else {
			if p.currentTok.Type == op {
				return true
			}
		}
	}

	return false
}

func (p *Parser) expectLBraceWithOptionalNewline(res *ParseResult) error {
	p.skipWhitespace(res)

	if p.currentTok.Type != constants.TT_LBRACE {
		return errors.NewInvalidSyntaxError(
			p.currentTok.PosStart, p.currentTok.PosEnd,
			"Expected '{'",
		)
	}

	res.RegisterAdvancement()
	p.advance()

	return nil
}
