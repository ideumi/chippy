/*
 *
 * Modena - internal/bytecode/chunk.go
 *
 */

package bytecode

import (
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"encoding/binary"
)

type Span struct {
	Start *errors.Position
	End   *errors.Position
}

type Chunk struct {
	Code      []byte
	Constants []values.Value
	Names     []string
	Functions []*FunctionTemplate

	// NumSlots is how many local variables this chunk needs room for and ...
	NumSlots int

	// ...LocalNames is the name of each one.
	LocalNames []string

	Spans []Span

	// indexSpans records where each part of an index expression was written,
	// so an error can point at just the collection, the index, or the value
	// instead of the whole expression. It only improves error messages and
	// is not written into a bytecode file.
	indexSpans map[int]IndexSpanSet
}

type IndexSpanSet struct {
	Coll Span
	Idx  Span
	Val  Span
}

func NewChunk() *Chunk {
	return &Chunk{}
}

func (c *Chunk) SetIndexSpans(offset int, spans IndexSpanSet) {
	if c.indexSpans == nil {
		c.indexSpans = map[int]IndexSpanSet{}
	}

	c.indexSpans[offset] = spans
}

func (c *Chunk) IndexSpansAt(offset int) IndexSpanSet {
	if spans, ok := c.indexSpans[offset]; ok {
		return spans
	}

	fallback := c.Spans[offset]

	return IndexSpanSet{Coll: fallback, Idx: fallback, Val: fallback}
}

func (c *Chunk) Emit(op Op, span Span, operands ...uint32) int {
	offset := len(c.Code)
	c.Code = append(c.Code, byte(op))
	c.Spans = append(c.Spans, span)

	for _, operand := range operands {
		c.Code = binary.LittleEndian.AppendUint32(c.Code, operand)
		c.Spans = append(c.Spans, span, span, span, span)
	}

	return offset
}

func (c *Chunk) AddConstant(value values.Value) uint32 {
	c.Constants = append(c.Constants, value)

	return uint32(len(c.Constants) - 1)
}

func (c *Chunk) AddName(name string) uint32 {
	c.Names = append(c.Names, name)

	return uint32(len(c.Names) - 1)
}

func (c *Chunk) AddFunction(fn *FunctionTemplate) uint32 {
	c.Functions = append(c.Functions, fn)

	return uint32(len(c.Functions) - 1)
}

func ReadU32(code []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(code[offset:])
}

// EmitJump writes a jump whose destination is not known yet, and returns the
// position of the empty destination so PatchJump can fill it in later.
func (c *Chunk) EmitJump(op Op, span Span) int {
	c.Emit(op, span, 0)

	return len(c.Code) - 4
}

func (c *Chunk) PatchJump(operandOffset int) {
	c.PatchJumpTo(operandOffset, len(c.Code))
}

func (c *Chunk) PatchJumpTo(operandOffset, target int) {
	binary.LittleEndian.PutUint32(c.Code[operandOffset:], uint32(target))
}
