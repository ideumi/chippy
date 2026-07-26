/*
 *
 * Modena - internal/bytecode/chunk.go
 *
 */

package bytecode

import (
	"chip-go/internal/errors"
	"chip-go/internal/globals"
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

	spanIndex   []uint32
	sharedSpans []Span
	seenSpans   map[Span]uint32

	// Only improves error messages, not written into bytecode
	indexSpans map[int]IndexSpanSet

	globalSlots []int
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

	fallback := c.SpanAt(offset)

	return IndexSpanSet{Coll: fallback, Idx: fallback, Val: fallback}
}

func (c *Chunk) shareSpan(span Span) uint32 {
	if c.seenSpans == nil {
		c.seenSpans = map[Span]uint32{}
	}

	if index, ok := c.seenSpans[span]; ok {
		return index
	}

	index := uint32(len(c.sharedSpans))
	c.sharedSpans = append(c.sharedSpans, span)
	c.seenSpans[span] = index

	return index
}

func (c *Chunk) SpanAt(offset int) Span {
	if offset < 0 || offset >= len(c.spanIndex) {
		return Span{}
	}

	return c.sharedSpans[c.spanIndex[offset]]
}

func (c *Chunk) SpanCount() int {
	return len(c.spanIndex)
}

func (c *Chunk) FinishBuilding() {
	c.seenSpans = nil
}

func (c *Chunk) Emit(op Op, span Span, operands ...uint32) int {
	offset := len(c.Code)
	index := c.shareSpan(span)

	c.Code = append(c.Code, byte(op))
	c.spanIndex = append(c.spanIndex, index)

	for _, operand := range operands {
		c.Code = binary.LittleEndian.AppendUint32(c.Code, operand)
		c.spanIndex = append(c.spanIndex, index, index, index, index)
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

// LinkGlobals resolves every global name in this chunk and in its nested
// function chunks to its schema slot.
func (c *Chunk) LinkGlobals(schema *globals.Schema) {
	c.globalSlots = make([]int, len(c.Names))

	for i, name := range c.Names {
		c.globalSlots[i] = schema.Intern(name)
	}

	for _, fn := range c.Functions {
		fn.Chunk.LinkGlobals(schema)
	}
}

func (c *Chunk) GlobalSlot(nameIndex int) int {
	return c.globalSlots[nameIndex]
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
