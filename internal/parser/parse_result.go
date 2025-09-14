/*
 *
 * RR2 - internal/parser/parse_result.go
 *
 */

package parser

import (
	"chip-go/internal/ast"
)

type ParseResult struct {
	error                      error
	node                       ast.Node
	lastRegisteredAdvanceCount int
	advanceCount               int
	toReverseCount             int
}

func NewParseResult() *ParseResult {
	return &ParseResult{
		lastRegisteredAdvanceCount: 0,
		advanceCount:               0,
		toReverseCount:             0,
	}
}

func (pr *ParseResult) RegisterAdvancement() {
	pr.lastRegisteredAdvanceCount = 1
	pr.advanceCount++
}

func (pr *ParseResult) Register(res *ParseResult) ast.Node {
	pr.lastRegisteredAdvanceCount = res.advanceCount
	pr.advanceCount += res.advanceCount
	if res.error != nil {
		pr.error = res.error
	}

	return res.node
}

func (pr *ParseResult) TryRegister(res *ParseResult) ast.Node {
	if res.error != nil {
		pr.toReverseCount = res.advanceCount
		return nil
	}

	return pr.Register(res)
}

func (pr *ParseResult) Success(node ast.Node) *ParseResult {
	pr.node = node

	return pr
}

func (pr *ParseResult) Failure(err error) *ParseResult {
	if pr.error == nil || pr.lastRegisteredAdvanceCount == 0 {
		pr.error = err
	}

	return pr
}

func (pr *ParseResult) GetError() error {
	return pr.error
}

func (pr *ParseResult) GetNode() ast.Node {
	return pr.node
}

func (pr *ParseResult) GetAdvanceCount() int {
	return pr.advanceCount
}

func (pr *ParseResult) GetToReverseCount() int {
	return pr.toReverseCount
}
