/*
 *
 * Modena - internal/compiler/compile.go
 *
 */

package compiler

import (
	"chip-go/internal/ast"
	"chip-go/internal/bytecode"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/lexer"
	"chip-go/internal/values"
	"fmt"
)

type Compiler struct {
	chunk      *bytecode.Chunk
	loops      []*loopContext
	scopes     []*funcScope
	blockDepth int
}

type loopContext struct {
	continueJumps []int
	breakJumps    []int
}

// emitBind leaves the value where it is rather than consuming it, so an assignment
// still has a result.
func (c *Compiler) emitBind(name string, span bytecode.Span) {
	if c.atModuleLevel() {
		c.chunk.Emit(bytecode.OpDefineGlobal, span, c.chunk.AddName(name))

		return
	}

	c.chunk.Emit(bytecode.OpSetLocal, span, uint32(c.addLocal(name)))
}

// atModuleLevel is true only at the top of a file, outside every block. A variable
// declared there becomes a module global, which is how a loaded file shares its
// names with the file that loaded it. Anywhere else it is a local.
func (c *Compiler) atModuleLevel() bool {
	return len(c.scopes) == 1 && c.blockDepth == 0
}

func NewCompiler() *Compiler {
	return &Compiler{chunk: bytecode.NewChunk()}
}

func (c *Compiler) Compile(node ast.Node) (*bytecode.Chunk, error) {
	c.scopes = append(c.scopes, &funcScope{})

	if err := c.compileProgram(node); err != nil {
		return nil, err
	}

	root := c.currentScope()
	c.chunk.NumSlots = root.slots
	c.chunk.LocalNames = root.names

	return c.chunk, nil
}

func (c *Compiler) compileProgram(node ast.Node) error {
	block, ok := node.(*ast.BlockNode)

	if !ok {
		return c.compile(node)
	}

	if len(block.ElementNodes) == 0 {
		c.emitConstant(values.NewNumber(constants.NUM_NUL), spanOf(block))

		return nil
	}

	return c.compileSequence(block)
}

func (c *Compiler) compile(node ast.Node) error {
	switch typed := node.(type) {
	case *ast.BlockNode:
		return c.compileBlock(typed)
	case *ast.NumberNode:
		return c.compileNumber(typed)
	case *ast.StringNode:
		return c.compileString(typed)
	case *ast.ListNode:
		return c.compileList(typed)
	case *ast.ByteArrayNode:
		return c.compileByteArray(typed)
	case *ast.MapNode:
		return c.compileMap(typed)
	case *ast.IndexAccessNode:
		return c.compileIndexAccess(typed)
	case *ast.IndexAssignNode:
		return c.compileIndexAssign(typed)
	case *ast.VarAccessNode:
		return c.compileVarAccess(typed)
	case *ast.VarAssignNode:
		return c.compileVarAssign(typed)
	case *ast.VarUpdateNode:
		return c.compileVarUpdate(typed)
	case *ast.BinOpNode:
		return c.compileBinOp(typed)
	case *ast.UnaryOpNode:
		return c.compileUnaryOp(typed)
	case *ast.IfNode:
		return c.compileIf(typed)
	case *ast.WhileNode:
		return c.compileWhile(typed)
	case *ast.ForNode:
		return c.compileFor(typed)
	case *ast.BreakNode:
		return c.compileBreak(typed)
	case *ast.ContinueNode:
		return c.compileContinue(typed)
	case *ast.FuncDefNode:
		return c.compileFuncDef(typed)
	case *ast.CallNode:
		return c.compileCall(typed)
	case *ast.ReturnNode:
		return c.compileReturn(typed)
	default:
		return bytecode.ModenaError("no compile rule for node type %T", node)
	}
}

func (c *Compiler) compileBlock(node *ast.BlockNode) error {
	if len(node.ElementNodes) == 0 {
		c.emitConstant(values.NewNumber(constants.NUM_NUL), spanOf(node))

		return nil
	}

	mark := c.beginScope()
	err := c.compileSequence(node)
	c.endScope(mark)

	return err
}

// Every result but the last is thrown away, so the value of a block is the value
// of its final statement.
func (c *Compiler) compileSequence(node *ast.BlockNode) error {
	for idx, statement := range node.ElementNodes {
		if err := c.compile(statement); err != nil {
			return err
		}

		if idx < len(node.ElementNodes)-1 {
			c.chunk.Emit(bytecode.OpDiscard, spanOf(statement))
		}
	}

	return nil
}

func (c *Compiler) compileNumber(node *ast.NumberNode) error {
	num, err := numberFromToken(node)

	if err != nil {
		return err
	}

	c.emitConstant(num, spanOf(node))

	return nil
}

func (c *Compiler) compileString(node *ast.StringNode) error {
	c.emitConstant(values.NewString(node.Token.Value.(string)), spanOf(node))

	return nil
}

func (c *Compiler) compileList(node *ast.ListNode) error {
	for _, element := range node.ElementNodes {
		if err := c.compile(element); err != nil {
			return err
		}
	}

	c.chunk.Emit(bytecode.OpBuildList, spanOf(node), uint32(len(node.ElementNodes)))

	return nil
}

func (c *Compiler) compileByteArray(node *ast.ByteArrayNode) error {
	for _, element := range node.ElementNodes {
		if err := c.compile(element); err != nil {
			return err
		}
	}

	c.chunk.Emit(bytecode.OpBuildBytes, spanOf(node), uint32(len(node.ElementNodes)))

	return nil
}

func (c *Compiler) compileMap(node *ast.MapNode) error {
	for i, key := range node.KeyNodes {
		if err := c.compile(key); err != nil {
			return err
		}

		if err := c.compile(node.ValueNodes[i]); err != nil {
			return err
		}
	}

	c.chunk.Emit(bytecode.OpBuildMap, spanOf(node), uint32(len(node.KeyNodes)))

	return nil
}

func (c *Compiler) compileIndexAccess(node *ast.IndexAccessNode) error {
	if err := c.compile(node.CollectionNode); err != nil {
		return err
	}

	if err := c.compile(node.IndexNode); err != nil {
		return err
	}

	offset := c.chunk.Emit(bytecode.OpIndexGet, spanOf(node))
	c.chunk.SetIndexSpans(offset, bytecode.IndexSpanSet{Coll: spanOf(node.CollectionNode), Idx: spanOf(node.IndexNode)})

	return nil
}

func (c *Compiler) compileIndexAssign(node *ast.IndexAssignNode) error {
	if err := c.compile(node.CollectionNode); err != nil {
		return err
	}

	if err := c.compile(node.IndexNode); err != nil {
		return err
	}

	if err := c.compile(node.ValueNode); err != nil {
		return err
	}

	offset := c.chunk.Emit(bytecode.OpIndexSet, spanOf(node))
	c.chunk.SetIndexSpans(offset, bytecode.IndexSpanSet{Coll: spanOf(node.CollectionNode), Idx: spanOf(node.IndexNode), Val: spanOf(node.ValueNode)})

	return nil
}

func (c *Compiler) compileVarAccess(node *ast.VarAccessNode) error {
	name := node.VarNameToken.Value.(string)

	if slot := c.resolveLocal(name); slot >= 0 {
		c.chunk.Emit(bytecode.OpGetLocal, spanOf(node), uint32(slot))

		return nil
	}

	if len(c.scopes) > 1 {
		if up := c.resolveUpvalue(len(c.scopes)-1, name); up >= 0 {
			c.chunk.Emit(bytecode.OpGetUpvalue, spanOf(node), uint32(up))

			return nil
		}
	}

	c.chunk.Emit(bytecode.OpGetGlobal, spanOf(node), c.chunk.AddName(name))

	return nil
}

func (c *Compiler) compileVarAssign(node *ast.VarAssignNode) error {
	if err := c.compile(node.ValueNode); err != nil {
		return err
	}

	c.emitBind(node.VarNameToken.Value.(string), spanOf(node))

	return nil
}

func (c *Compiler) compileVarUpdate(node *ast.VarUpdateNode) error {
	if err := c.compile(node.ValueNode); err != nil {
		return err
	}

	name := node.VarNameToken.Value.(string)

	if slot := c.resolveLocal(name); slot >= 0 {
		c.chunk.Emit(bytecode.OpSetLocal, spanOf(node), uint32(slot))

		return nil
	}

	if len(c.scopes) > 1 {
		if up := c.resolveUpvalue(len(c.scopes)-1, name); up >= 0 {
			c.chunk.Emit(bytecode.OpSetUpvalue, spanOf(node), uint32(up))

			return nil
		}
	}

	c.chunk.Emit(bytecode.OpSetGlobal, spanOf(node), c.chunk.AddName(name))

	return nil
}

func (c *Compiler) compileBinOp(node *ast.BinOpNode) error {
	if isLogicalKeyword(node.OpToken) {
		return c.compileLogical(node)
	}

	if err := c.compile(node.LeftNode); err != nil {
		return err
	}

	if err := c.compile(node.RightNode); err != nil {
		return err
	}

	op, ok := binOpForToken(node.OpToken)

	if !ok {
		return bytecode.ModenaError("unsupported binary operator %v", node.OpToken.Type)
	}

	c.chunk.Emit(op, spanOf(node))

	return nil
}

func (c *Compiler) compileUnaryOp(node *ast.UnaryOpNode) error {
	if err := c.compile(node.Node); err != nil {
		return err
	}

	span := spanOf(node)

	switch node.OpToken.Type {
	case constants.TT_MINUS:
		c.chunk.Emit(bytecode.OpNeg, span)
	case constants.TT_PLUS:
		c.chunk.Emit(bytecode.OpPos, span)
	case constants.TT_KEYWORD:
		switch node.OpToken.Value {
		case "not":
			c.chunk.Emit(bytecode.OpNot, span)
		case "bnot":
			c.chunk.Emit(bytecode.OpBNot, span)
		default:
			return bytecode.ModenaError("unsupported unary keyword %v", node.OpToken.Value)
		}
	default:
		return bytecode.ModenaError("unsupported unary operator %v", node.OpToken.Type)
	}

	return nil
}

func (c *Compiler) compileIf(node *ast.IfNode) error {
	var endJumps []int

	for _, ifCase := range node.Cases {
		if err := c.compile(ifCase.Condition); err != nil {
			return err
		}

		next := c.chunk.EmitJump(bytecode.OpJumpIfFalse, spanOf(node))

		// OpJumpIfFalse does not take the condition off the stack, so
		// it is dropped on the path that runs the body and on the path
		// that skips it.
		c.chunk.Emit(bytecode.OpDiscard, spanOf(node))

		if err := c.compileIfBody(ifCase); err != nil {
			return err
		}

		endJumps = append(endJumps, c.chunk.EmitJump(bytecode.OpJump, spanOf(node)))

		c.chunk.PatchJump(next)
		c.chunk.Emit(bytecode.OpDiscard, spanOf(node))
	}

	if node.ElseCase != nil {
		if err := c.compile(node.ElseCase); err != nil {
			return err
		}
	} else {
		c.emitConstant(values.NewNumber(constants.NUM_NUL), spanOf(node))
	}

	for _, jump := range endJumps {
		c.chunk.PatchJump(jump)
	}

	return nil
}

func (c *Compiler) compileIfBody(ifCase ast.IfCase) error {
	if err := c.compile(ifCase.Body); err != nil {
		return err
	}

	if ifCase.ShouldReturnNull {
		c.chunk.Emit(bytecode.OpDiscard, spanOf(ifCase.Body))
		c.emitConstant(values.NewNumber(constants.NUM_NUL), spanOf(ifCase.Body))
	}

	return nil
}

// 'and' and 'or' return one of their two operands, never true or false. 'and'
// keeps the left value when it is false and runs the right side otherwise. 'or'
// keeps the left value when it is true and runs the right side otherwise. The
// side that wins is the one left on the stack.
func (c *Compiler) compileLogical(node *ast.BinOpNode) error {
	if err := c.compile(node.LeftNode); err != nil {
		return err
	}

	if node.OpToken.Value == "and" {
		skip := c.chunk.EmitJump(bytecode.OpJumpIfFalse, spanOf(node))
		c.chunk.Emit(bytecode.OpDiscard, spanOf(node))

		if err := c.compile(node.RightNode); err != nil {
			return err
		}

		c.chunk.PatchJump(skip)

		return nil
	}

	evalRight := c.chunk.EmitJump(bytecode.OpJumpIfFalse, spanOf(node))
	end := c.chunk.EmitJump(bytecode.OpJump, spanOf(node))

	c.chunk.PatchJump(evalRight)
	c.chunk.Emit(bytecode.OpDiscard, spanOf(node))

	if err := c.compile(node.RightNode); err != nil {
		return err
	}

	c.chunk.PatchJump(end)

	return nil
}

func (c *Compiler) compileWhile(node *ast.WhileNode) error {
	start := len(c.chunk.Code)

	if err := c.compile(node.ConditionNode); err != nil {
		return err
	}

	exit := c.chunk.EmitJump(bytecode.OpJumpIfFalse, spanOf(node))
	c.chunk.Emit(bytecode.OpDiscard, spanOf(node))

	c.loops = append(c.loops, &loopContext{})

	if err := c.compile(node.BodyNode); err != nil {
		return err
	}

	c.chunk.Emit(bytecode.OpDiscard, spanOf(node))
	c.chunk.Emit(bytecode.OpJump, spanOf(node), uint32(start))

	loop := c.loops[len(c.loops)-1]
	c.loops = c.loops[:len(c.loops)-1]

	c.chunk.PatchJump(exit)
	c.chunk.Emit(bytecode.OpDiscard, spanOf(node))

	for _, brk := range loop.breakJumps {
		c.chunk.PatchJump(brk)
	}

	for _, cnt := range loop.continueJumps {
		c.chunk.PatchJumpTo(cnt, start)
	}

	c.emitConstant(values.NewNumber(constants.NUM_NUL), spanOf(node))

	return nil
}

func (c *Compiler) compileFor(node *ast.ForNode) error {
	if err := c.compile(node.StartValueNode); err != nil {
		return err
	}

	if err := c.compile(node.EndValueNode); err != nil {
		return err
	}

	if node.StepValueNode != nil {
		if err := c.compile(node.StepValueNode); err != nil {
			return err
		}
	} else {
		c.emitConstant(values.NewNumber(1), spanOf(node))
	}

	c.chunk.Emit(bytecode.OpForBegin, spanOf(node))

	// The loop variable belongs to the loop alone, so it is declared inside
	// a new block that ends when the loop does.
	mark := c.beginScope()

	testPos := len(c.chunk.Code)
	exit := c.chunk.EmitJump(bytecode.OpForTest, spanOf(node))

	// OpForTest leaves the loop counter on the stack. It is stored into the
	// loop variable here, and the copy left behind is dropped.
	c.emitBind(node.VarNameToken.Value.(string), spanOf(node))
	c.chunk.Emit(bytecode.OpDiscard, spanOf(node))

	c.loops = append(c.loops, &loopContext{})

	if err := c.compile(node.BodyNode); err != nil {
		return err
	}

	c.chunk.Emit(bytecode.OpDiscard, spanOf(node))

	loop := c.loops[len(c.loops)-1]
	c.loops = c.loops[:len(c.loops)-1]

	next := len(c.chunk.Code)
	c.chunk.Emit(bytecode.OpForNext, spanOf(node), uint32(testPos))

	end := len(c.chunk.Code)
	c.chunk.PatchJumpTo(exit, end)

	for _, brk := range loop.breakJumps {
		c.chunk.PatchJumpTo(brk, end)
	}

	c.chunk.Emit(bytecode.OpForEnd, spanOf(node))

	for _, cnt := range loop.continueJumps {
		c.chunk.PatchJumpTo(cnt, next)
	}

	c.endScope(mark)

	c.emitConstant(values.NewNumber(constants.NUM_NUL), spanOf(node))

	return nil
}

func (c *Compiler) compileBreak(node *ast.BreakNode) error {
	if len(c.loops) == 0 {
		return errors.NewRTError(node.GetPosStart(), node.GetPosEnd(), "'break' outside a loop")
	}

	loop := c.loops[len(c.loops)-1]
	loop.breakJumps = append(loop.breakJumps, c.chunk.EmitJump(bytecode.OpJump, spanOf(node)))

	return nil
}

func (c *Compiler) compileContinue(node *ast.ContinueNode) error {
	if len(c.loops) == 0 {
		return errors.NewRTError(node.GetPosStart(), node.GetPosEnd(), "'continue' outside a loop")
	}

	loop := c.loops[len(c.loops)-1]
	loop.continueJumps = append(loop.continueJumps, c.chunk.EmitJump(bytecode.OpJump, spanOf(node)))

	return nil
}

func (c *Compiler) compileFuncDef(node *ast.FuncDefNode) error {
	name := ""

	if node.VarNameToken != nil {
		name = node.VarNameToken.Value.(string)
	}

	argNames := make([]string, len(node.ArgNameTokens))

	for i, token := range node.ArgNameTokens {
		argNames[i] = token.Value.(string)
	}

	// A named function is given its name before its body is compiled, so
	// the body can call the function itself. A function written further down
	// the file does not exist yet, so two functions that call each other
	// only work in one order.
	if name != "" && !c.atModuleLevel() {
		c.addLocal(name)
	}

	parentChunk := c.chunk
	c.chunk = bytecode.NewChunk()
	c.scopes = append(c.scopes, &funcScope{})

	for _, arg := range argNames {
		c.addLocal(arg)
	}

	err := c.compile(node.BodyNode)

	if err == nil {
		// A function does not use the value of its body, so it is dropped
		// and a return of null added for a body that ends without a return
		// of its own.
		c.chunk.Emit(bytecode.OpDiscard, spanOf(node.BodyNode))
		c.emitConstant(values.NewNumber(constants.NUM_NUL), spanOf(node))
		c.chunk.Emit(bytecode.OpReturn, spanOf(node))
	}

	scope := c.scopes[len(c.scopes)-1]
	fnChunk := c.chunk
	fnChunk.NumSlots = scope.slots
	fnChunk.LocalNames = scope.names
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.chunk = parentChunk

	if err != nil {
		return err
	}

	fn := &bytecode.FunctionTemplate{Name: name, Chunk: fnChunk, Arity: len(argNames), Upvalues: scope.upvalues}
	c.chunk.Emit(bytecode.OpClosure, spanOf(node), c.chunk.AddFunction(fn))

	if node.VarNameToken != nil {
		c.bindFunc(name, spanOf(node))
	}

	return nil
}

func (c *Compiler) bindFunc(name string, span bytecode.Span) {
	if c.atModuleLevel() {
		c.chunk.Emit(bytecode.OpDefineGlobal, span, c.chunk.AddName(name))

		return
	}

	slot := c.resolveLocal(name)

	if slot < 0 {
		slot = c.addLocal(name)
	}

	c.chunk.Emit(bytecode.OpSetLocal, span, uint32(slot))
}

func (c *Compiler) compileCall(node *ast.CallNode) error {
	if err := c.compile(node.NodeToCall); err != nil {
		return err
	}

	for _, argNode := range node.ArgNodes {
		if err := c.compile(argNode); err != nil {
			return err
		}
	}

	c.chunk.Emit(bytecode.OpCall, spanOf(node), uint32(len(node.ArgNodes)))

	return nil
}

func (c *Compiler) compileReturn(node *ast.ReturnNode) error {
	// Only a function body adds a scope on top of the program's own, so with
	// just the one this return is outside any function, block or no block.

	if len(c.scopes) <= 1 {
		return errors.NewRTError(node.GetPosStart(), node.GetPosEnd(), "'return' outside a function")
	}

	if node.NodeToReturn != nil {
		if err := c.compile(node.NodeToReturn); err != nil {
			return err
		}
	} else {
		c.emitConstant(values.NewNumber(constants.NUM_NUL), spanOf(node))
	}

	c.chunk.Emit(bytecode.OpReturn, spanOf(node))

	return nil
}

func (c *Compiler) emitConstant(value values.Value, span bytecode.Span) {
	index := c.chunk.AddConstant(value)
	c.chunk.Emit(bytecode.OpConstant, span, index)
}

func spanOf(node ast.Node) bytecode.Span {
	return bytecode.Span{Start: node.GetPosStart(), End: node.GetPosEnd()}
}

func numberFromToken(node *ast.NumberNode) (*values.Number, error) {
	switch value := node.Token.Value.(type) {
	case int64:
		return values.NewNumber(value), nil
	case int:
		return values.NewNumber(int64(value)), nil
	case float64:
		num, err := values.NewNumberFromFloat(value)

		if err != nil {
			return nil, errors.NewRTError(node.PosStart, node.PosEnd, err.Error())
		}

		return num, nil
	default:
		return nil, errors.NewRTError(
			node.PosStart, node.PosEnd,
			fmt.Sprintf("Invalid number value: %v", node.Token.Value))
	}
}

func isLogicalKeyword(token *lexer.Token) bool {
	return token.Type == constants.TT_KEYWORD && (token.Value == "and" || token.Value == "or")
}

func binOpForToken(token *lexer.Token) (bytecode.Op, bool) {
	switch token.Type {
	case constants.TT_PLUS:
		return bytecode.OpAdd, true
	case constants.TT_MINUS:
		return bytecode.OpSub, true
	case constants.TT_MUL:
		return bytecode.OpMul, true
	case constants.TT_DIV:
		return bytecode.OpDiv, true
	case constants.TT_MOD:
		return bytecode.OpMod, true
	case constants.TT_POW:
		return bytecode.OpPow, true
	case constants.TT_LSHIFT:
		return bytecode.OpLShift, true
	case constants.TT_RSHIFT:
		return bytecode.OpRShift, true
	case constants.TT_EE:
		return bytecode.OpEq, true
	case constants.TT_NE:
		return bytecode.OpNe, true
	case constants.TT_LT:
		return bytecode.OpLt, true
	case constants.TT_GT:
		return bytecode.OpGt, true
	case constants.TT_LTE:
		return bytecode.OpLte, true
	case constants.TT_GTE:
		return bytecode.OpGte, true
	case constants.TT_KEYWORD:
		switch token.Value {
		case "xor":
			return bytecode.OpXor, true
		case "band":
			return bytecode.OpBAnd, true
		case "bor":
			return bytecode.OpBOr, true
		case "bxor":
			return bytecode.OpBXor, true
		}
	}

	return 0, false
}
