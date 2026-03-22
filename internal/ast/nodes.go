/*
 *
 * RR2 - internal/ast/nodes.go
 *
 */

package ast

import (
	"chip-go/internal/errors"
	"chip-go/internal/lexer"
	"fmt"
	"strings"
)

type NumberNode struct {
	*BaseNode
	Token *lexer.Token
}

func NewNumberNode(token *lexer.Token) *NumberNode {
	return &NumberNode{
		BaseNode: NewBaseNode(token.PosStart, token.PosEnd),
		Token:    token,
	}
}

func (n *NumberNode) String() string {
	return fmt.Sprintf("%v", n.Token.Value)
}

type StringNode struct {
	*BaseNode
	Token *lexer.Token
}

func NewStringNode(token *lexer.Token) *StringNode {
	return &StringNode{
		BaseNode: NewBaseNode(token.PosStart, token.PosEnd),
		Token:    token,
	}
}

func (n *StringNode) String() string {
	return fmt.Sprintf("\"%s\"", n.Token.Value)
}

type ListNode struct {
	*BaseNode
	ElementNodes []Node
}

func NewListNode(elementNodes []Node, posStart, posEnd *errors.Position) *ListNode {
	return &ListNode{
		BaseNode:     NewBaseNode(posStart, posEnd),
		ElementNodes: elementNodes,
	}
}

func (n *ListNode) String() string {
	elements := make([]string, len(n.ElementNodes))

	for i, node := range n.ElementNodes {
		elements[i] = node.String()
	}

	return "[" + strings.Join(elements, ", ") + "]"
}

type ByteArrayNode struct {
	*BaseNode
	ElementNodes []Node
}

func NewByteArrayNode(elementNodes []Node, posStart, posEnd *errors.Position) *ByteArrayNode {
	return &ByteArrayNode{
		BaseNode:     NewBaseNode(posStart, posEnd),
		ElementNodes: elementNodes,
	}
}

func (n *ByteArrayNode) String() string {
	elements := make([]string, len(n.ElementNodes))

	for i, node := range n.ElementNodes {
		elements[i] = node.String()
	}

	return "b[" + strings.Join(elements, ", ") + "]"
}

type MapNode struct {
	*BaseNode
	KeyNodes   []Node
	ValueNodes []Node
}

func NewMapNode(posStart, posEnd *errors.Position, keyNodes []Node, valueNodes []Node) *MapNode {
	return &MapNode{
		BaseNode:   NewBaseNode(posStart, posEnd),
		KeyNodes:   keyNodes,
		ValueNodes: valueNodes,
	}
}

func (n *MapNode) String() string {
	if len(n.KeyNodes) == 0 {
		return "m[]"
	}

	pairs := make([]string, len(n.KeyNodes))

	for i, k := range n.KeyNodes {
		pairs[i] = k.String() + ": " + n.ValueNodes[i].String()
	}

	return "m[" + strings.Join(pairs, ", ") + "]"
}

type VarAccessNode struct {
	*BaseNode
	VarNameToken *lexer.Token
}

func NewVarAccessNode(varNameToken *lexer.Token) *VarAccessNode {
	return &VarAccessNode{
		BaseNode:     NewBaseNode(varNameToken.PosStart, varNameToken.PosEnd),
		VarNameToken: varNameToken,
	}
}

func (n *VarAccessNode) String() string {
	return fmt.Sprintf("%s", n.VarNameToken.Value)
}

type VarAssignNode struct {
	*BaseNode
	VarNameToken *lexer.Token
	ValueNode    Node
}

func NewVarAssignNode(varNameToken *lexer.Token, valueNode Node) *VarAssignNode {
	return &VarAssignNode{
		BaseNode:     NewBaseNode(varNameToken.PosStart, valueNode.GetPosEnd()),
		VarNameToken: varNameToken,
		ValueNode:    valueNode,
	}
}

func (n *VarAssignNode) String() string {
	return fmt.Sprintf("%s = %s", n.VarNameToken.Value, n.ValueNode.String())
}

type VarUpdateNode struct {
	*BaseNode
	VarNameToken *lexer.Token
	ValueNode    Node
}

func NewVarUpdateNode(varNameToken *lexer.Token, valueNode Node) *VarUpdateNode {
	return &VarUpdateNode{
		BaseNode:     NewBaseNode(varNameToken.PosStart, valueNode.GetPosEnd()),
		VarNameToken: varNameToken,
		ValueNode:    valueNode,
	}
}

func (n *VarUpdateNode) String() string {
	return fmt.Sprintf("%s = %s", n.VarNameToken.Value, n.ValueNode.String())
}

type BinOpNode struct {
	*BaseNode
	LeftNode  Node
	OpToken   *lexer.Token
	RightNode Node
}

func NewBinOpNode(leftNode Node, opToken *lexer.Token, rightNode Node) *BinOpNode {
	return &BinOpNode{
		BaseNode:  NewBaseNode(leftNode.GetPosStart(), rightNode.GetPosEnd()),
		LeftNode:  leftNode,
		OpToken:   opToken,
		RightNode: rightNode,
	}
}

func (n *BinOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.LeftNode.String(), n.OpToken.Type, n.RightNode.String())
}

type UnaryOpNode struct {
	*BaseNode
	OpToken *lexer.Token
	Node    Node
}

func NewUnaryOpNode(opToken *lexer.Token, node Node) *UnaryOpNode {
	return &UnaryOpNode{
		BaseNode: NewBaseNode(opToken.PosStart, node.GetPosEnd()),
		OpToken:  opToken,
		Node:     node,
	}
}

func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("(%s%s)", n.OpToken.Type, n.Node.String())
}

type IfNode struct {
	*BaseNode
	Cases    []IfCase
	ElseCase Node
}

type IfCase struct {
	Condition        Node
	Body             Node
	ShouldReturnNull bool
}

func NewIfNode(cases []IfCase, elseCase Node) *IfNode {
	var posStart, posEnd *errors.Position
	if len(cases) > 0 {
		posStart = cases[0].Condition.GetPosStart()
		posEnd = cases[len(cases)-1].Body.GetPosEnd()
	}
	if elseCase != nil {
		posEnd = elseCase.GetPosEnd()
	}

	return &IfNode{
		BaseNode: NewBaseNode(posStart, posEnd),
		Cases:    cases,
		ElseCase: elseCase,
	}
}

func (n *IfNode) String() string {
	return "IF"
}

type ForNode struct {
	*BaseNode
	VarNameToken     *lexer.Token
	StartValueNode   Node
	EndValueNode     Node
	StepValueNode    Node
	BodyNode         Node
	ShouldReturnNull bool
}

func NewForNode(varNameToken *lexer.Token, startValueNode, endValueNode, stepValueNode, bodyNode Node, shouldReturnNull bool) *ForNode {
	return &ForNode{
		BaseNode:         NewBaseNode(varNameToken.PosStart, bodyNode.GetPosEnd()),
		VarNameToken:     varNameToken,
		StartValueNode:   startValueNode,
		EndValueNode:     endValueNode,
		StepValueNode:    stepValueNode,
		BodyNode:         bodyNode,
		ShouldReturnNull: shouldReturnNull,
	}
}

func (n *ForNode) String() string {
	return "FOR"
}

type WhileNode struct {
	*BaseNode
	ConditionNode    Node
	BodyNode         Node
	ShouldReturnNull bool
}

func NewWhileNode(conditionNode, bodyNode Node, shouldReturnNull bool) *WhileNode {
	return &WhileNode{
		BaseNode:         NewBaseNode(conditionNode.GetPosStart(), bodyNode.GetPosEnd()),
		ConditionNode:    conditionNode,
		BodyNode:         bodyNode,
		ShouldReturnNull: shouldReturnNull,
	}
}

func (n *WhileNode) String() string {
	return "WHILE"
}

type FuncDefNode struct {
	*BaseNode
	VarNameToken     *lexer.Token
	ArgNameTokens    []*lexer.Token
	BodyNode         Node
	ShouldAutoReturn bool
}

func NewFuncDefNode(varNameToken *lexer.Token, argNameTokens []*lexer.Token, bodyNode Node, shouldAutoReturn bool) *FuncDefNode {
	var posStart *errors.Position
	if varNameToken != nil {
		posStart = varNameToken.PosStart
	} else if len(argNameTokens) > 0 {
		posStart = argNameTokens[0].PosStart
	} else {
		posStart = bodyNode.GetPosStart()
	}

	return &FuncDefNode{
		BaseNode:         NewBaseNode(posStart, bodyNode.GetPosEnd()),
		VarNameToken:     varNameToken,
		ArgNameTokens:    argNameTokens,
		BodyNode:         bodyNode,
		ShouldAutoReturn: shouldAutoReturn,
	}
}

func (n *FuncDefNode) String() string {
	return "FUNC"
}

type CallNode struct {
	*BaseNode
	NodeToCall Node
	ArgNodes   []Node
}

func NewCallNode(nodeToCall Node, argNodes []Node) *CallNode {
	var posEnd *errors.Position
	if len(argNodes) > 0 {
		posEnd = argNodes[len(argNodes)-1].GetPosEnd()
	} else {
		posEnd = nodeToCall.GetPosEnd()
	}

	return &CallNode{
		BaseNode:   NewBaseNode(nodeToCall.GetPosStart(), posEnd),
		NodeToCall: nodeToCall,
		ArgNodes:   argNodes,
	}
}

func (n *CallNode) String() string {
	return "CALL"
}

type ReturnNode struct {
	*BaseNode
	NodeToReturn Node
}

func NewReturnNode(nodeToReturn Node, posStart, posEnd *errors.Position) *ReturnNode {
	return &ReturnNode{
		BaseNode:     NewBaseNode(posStart, posEnd),
		NodeToReturn: nodeToReturn,
	}
}

func (n *ReturnNode) String() string {
	return "RETURN"
}

type ContinueNode struct {
	*BaseNode
}

func NewContinueNode(posStart, posEnd *errors.Position) *ContinueNode {
	return &ContinueNode{
		BaseNode: NewBaseNode(posStart, posEnd),
	}
}

func (n *ContinueNode) String() string {
	return "CONTINUE"
}

type BreakNode struct {
	*BaseNode
}

func NewBreakNode(posStart, posEnd *errors.Position) *BreakNode {
	return &BreakNode{
		BaseNode: NewBaseNode(posStart, posEnd),
	}
}

func (n *BreakNode) String() string {
	return "BREAK"
}

type IndexAccessNode struct {
	*BaseNode
	CollectionNode Node
	IndexNode      Node
}

func NewIndexAccessNode(collectionNode Node, indexNode Node, posEnd *errors.Position) *IndexAccessNode {
	return &IndexAccessNode{
		BaseNode:       NewBaseNode(collectionNode.GetPosStart(), posEnd),
		CollectionNode: collectionNode,
		IndexNode:      indexNode,
	}
}

func (n *IndexAccessNode) String() string {
	return fmt.Sprintf("%s[%s]", n.CollectionNode.String(), n.IndexNode.String())
}

type IndexAssignNode struct {
	*BaseNode
	CollectionNode Node
	IndexNode      Node
	ValueNode      Node
}

func NewIndexAssignNode(collectionNode Node, indexNode Node, valueNode Node) *IndexAssignNode {
	return &IndexAssignNode{
		BaseNode:       NewBaseNode(collectionNode.GetPosStart(), valueNode.GetPosEnd()),
		CollectionNode: collectionNode,
		IndexNode:      indexNode,
		ValueNode:      valueNode,
	}
}

func (n *IndexAssignNode) String() string {
	return fmt.Sprintf("%s[%s] = %s", n.CollectionNode.String(), n.IndexNode.String(), n.ValueNode.String())
}
