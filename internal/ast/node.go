/*
 *
 * Chippy - internal/ast/node.go
 *
 */

package ast

import "chip-go/internal/errors"

type Node interface {
	GetPosStart() *errors.Position
	GetPosEnd() *errors.Position
	String() string
}

type BaseNode struct {
	PosStart *errors.Position
	PosEnd   *errors.Position
}

func (n *BaseNode) GetPosStart() *errors.Position {
	return n.PosStart
}

func (n *BaseNode) GetPosEnd() *errors.Position {
	return n.PosEnd
}

func NewBaseNode(posStart, posEnd *errors.Position) *BaseNode {
	return &BaseNode{
		PosStart: posStart,
		PosEnd:   posEnd,
	}
}
