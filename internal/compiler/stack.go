/*
 *
 * Modena - internal/compiler/stack.go
 *
 */

package compiler

import (
	"chip-go/internal/bytecode"
	"chip-go/internal/errors"
)

func (c *Compiler) write(op bytecode.Op, span bytecode.Span, operands ...uint32) int {
	offset := c.chunk.Write(op, span, operands...)
	c.stackDepth += stackEffect(op, operands)

	return offset
}

func (c *Compiler) writeJump(op bytecode.Op, span bytecode.Span) int {
	offset := c.chunk.WriteJump(op, span)
	c.stackDepth += stackEffect(op, []uint32{0})

	return offset
}

func (c *Compiler) discardTo(depth int, span bytecode.Span) {
	for c.stackDepth > depth {
		c.write(bytecode.OpDiscard, span)
	}
}

func stackEffect(op bytecode.Op, operands []uint32) int {
	switch op {
	case bytecode.OpConstant, bytecode.OpGetGlobal, bytecode.OpGetLocal,
		bytecode.OpGetUpvalue, bytecode.OpClosure, bytecode.OpForTest:
		return 1

	case bytecode.OpDefineGlobal, bytecode.OpSetGlobal, bytecode.OpSetLocal,
		bytecode.OpSetUpvalue, bytecode.OpJump, bytecode.OpJumpIfFalse,
		bytecode.OpForNext, bytecode.OpForEnd,
		bytecode.OpNeg, bytecode.OpNot, bytecode.OpBNot:
		return 0

	case bytecode.OpDiscard, bytecode.OpReturn, bytecode.OpIndexGet,
		bytecode.OpAdd, bytecode.OpSub, bytecode.OpMul, bytecode.OpDiv,
		bytecode.OpMod, bytecode.OpPow, bytecode.OpLShift, bytecode.OpRShift,
		bytecode.OpEq, bytecode.OpNe, bytecode.OpLt, bytecode.OpGt,
		bytecode.OpLte, bytecode.OpGte, bytecode.OpXor, bytecode.OpBAnd,
		bytecode.OpBOr, bytecode.OpBXor:
		return -1

	case bytecode.OpIndexSet:
		return -2

	case bytecode.OpForBegin:
		return -3

	case bytecode.OpCall:
		return -int(operands[0])

	case bytecode.OpBuildList, bytecode.OpBuildBytes:
		return 1 - int(operands[0])

	case bytecode.OpBuildMap:
		return 1 - 2*int(operands[0])
	}

	errors.ModenaPanic("no stack effect for opcode %d", op)

	return 0
}
