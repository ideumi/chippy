/*
 *
 * Chippy - internal/bytecode/opcode.go
 *
 */

package bytecode

type Op byte

const (
	OpConstant Op = 0
	OpDiscard  Op = 1

	OpGetGlobal    Op = 2
	OpDefineGlobal Op = 3
	OpSetGlobal    Op = 4

	OpJump        Op = 5
	OpJumpIfFalse Op = 6

	OpGetLocal   Op = 7
	OpSetLocal   Op = 8
	OpGetUpvalue Op = 9
	OpSetUpvalue Op = 10
	OpClosure    Op = 11
	OpCall       Op = 12
	OpReturn     Op = 13

	OpForBegin Op = 14
	OpForTest  Op = 15
	OpForNext  Op = 16
	OpForEnd   Op = 17

	OpBuildList  Op = 18
	OpBuildBytes Op = 19
	OpBuildMap   Op = 20
	OpIndexGet   Op = 21
	OpIndexSet   Op = 22

	OpAdd    Op = 23
	OpSub    Op = 24
	OpMul    Op = 25
	OpDiv    Op = 26
	OpMod    Op = 27
	OpPow    Op = 28
	OpLShift Op = 29
	OpRShift Op = 30

	OpEq  Op = 31
	OpNe  Op = 32
	OpLt  Op = 33
	OpGt  Op = 34
	OpLte Op = 35
	OpGte Op = 36

	OpXor  Op = 37
	OpBAnd Op = 38
	OpBOr  Op = 39
	OpBXor Op = 40

	OpNeg  Op = 41
	OpNot  Op = 42
	OpBNot Op = 43
)

// OperandKind says what the value after an instruction means, so the disassembler
// can print it in a readable form. The VM already knows what to expect from each
// instruction and never looks at this.
type OperandKind int

const (
	OperandNone     OperandKind = 0
	OperandConstant OperandKind = 1
	OperandName     OperandKind = 2
	OperandJump     OperandKind = 3
	OperandInt      OperandKind = 4
	OperandFunction OperandKind = 5
)

type OpInfo struct {
	Name    string
	Operand OperandKind
}

var OpTable = map[Op]OpInfo{
	OpConstant:     {"OpConstant", OperandConstant},
	OpDiscard:      {"OpDiscard", OperandNone},
	OpGetGlobal:    {"OpGetGlobal", OperandName},
	OpDefineGlobal: {"OpDefineGlobal", OperandName},
	OpSetGlobal:    {"OpSetGlobal", OperandName},

	OpJump:        {"OpJump", OperandJump},
	OpJumpIfFalse: {"OpJumpIfFalse", OperandJump},

	OpGetLocal:   {"OpGetLocal", OperandInt},
	OpSetLocal:   {"OpSetLocal", OperandInt},
	OpGetUpvalue: {"OpGetUpvalue", OperandInt},
	OpSetUpvalue: {"OpSetUpvalue", OperandInt},
	OpClosure:    {"OpClosure", OperandFunction},
	OpCall:       {"OpCall", OperandInt},
	OpReturn:     {"OpReturn", OperandNone},

	OpForBegin: {"OpForBegin", OperandNone},
	OpForTest:  {"OpForTest", OperandJump},
	OpForNext:  {"OpForNext", OperandJump},
	OpForEnd:   {"OpForEnd", OperandNone},

	OpBuildList:  {"OpBuildList", OperandInt},
	OpBuildBytes: {"OpBuildBytes", OperandInt},
	OpBuildMap:   {"OpBuildMap", OperandInt},
	OpIndexGet:   {"OpIndexGet", OperandNone},
	OpIndexSet:   {"OpIndexSet", OperandNone},

	OpAdd:    {"OpAdd", OperandNone},
	OpSub:    {"OpSub", OperandNone},
	OpMul:    {"OpMul", OperandNone},
	OpDiv:    {"OpDiv", OperandNone},
	OpMod:    {"OpMod", OperandNone},
	OpPow:    {"OpPow", OperandNone},
	OpLShift: {"OpLShift", OperandNone},
	OpRShift: {"OpRShift", OperandNone},

	OpEq:  {"OpEq", OperandNone},
	OpNe:  {"OpNe", OperandNone},
	OpLt:  {"OpLt", OperandNone},
	OpGt:  {"OpGt", OperandNone},
	OpLte: {"OpLte", OperandNone},
	OpGte: {"OpGte", OperandNone},

	OpXor:  {"OpXor", OperandNone},
	OpBAnd: {"OpBAnd", OperandNone},
	OpBOr:  {"OpBOr", OperandNone},
	OpBXor: {"OpBXor", OperandNone},

	OpNeg:  {"OpNeg", OperandNone},
	OpNot:  {"OpNot", OperandNone},
	OpBNot: {"OpBNot", OperandNone},
}
