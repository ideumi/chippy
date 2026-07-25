/*
 *
 * Modena - internal/bytecode/serializer.go
 *
 */

package bytecode

import (
	"bytes"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"encoding/binary"
	"math"
)

var magic = []byte(constants.MODENA_MAGIC)

const (
	constTagInt    byte = 0
	constTagFloat  byte = 1
	constTagString byte = 2
)

func Serialize(chunk *Chunk, addShebang bool) []byte {
	out := &writer{}

	if addShebang {
		out.raw([]byte(constants.CHIPPY_SHEBANG + "\n"))
	}

	out.raw(magic)
	out.u32(constants.MODENA_FORMAT_VERSION)
	out.chunk(chunk)

	return out.buf
}

func Deserialize(data []byte) (*Chunk, error) {
	data = skipShebang(data)

	src := &reader{buf: data}

	if !src.expectMagic() {
		return nil, ModenaError("not a bytecode file")
	}

	version := src.u32()

	if version != constants.MODENA_FORMAT_VERSION {
		return nil, ModenaError("unsupported bytecode version %d, expected %d", version, constants.MODENA_FORMAT_VERSION)
	}

	chunk := src.chunk()

	if src.err != nil {
		return nil, src.err
	}

	return chunk, nil
}

func LooksLikeBytecode(data []byte) bool {
	data = skipShebang(data)

	return bytes.HasPrefix(data, magic)
}

func skipShebang(data []byte) []byte {
	if !bytes.HasPrefix(data, []byte("#!")) {
		return data
	}

	if newline := bytes.IndexByte(data, '\n'); newline >= 0 {
		return data[newline+1:]
	}

	return data
}

type writer struct {
	buf []byte
}

func (w *writer) raw(data []byte) {
	w.buf = append(w.buf, data...)
}

func (w *writer) u8(value byte) {
	w.buf = append(w.buf, value)
}

func (w *writer) u32(value uint32) {
	w.buf = binary.LittleEndian.AppendUint32(w.buf, value)
}

func (w *writer) u64(value uint64) {
	w.buf = binary.LittleEndian.AppendUint64(w.buf, value)
}

func (w *writer) str(value string) {
	w.u32(uint32(len(value)))
	w.raw([]byte(value))
}

func (w *writer) chunk(chunk *Chunk) {
	w.u32(uint32(len(chunk.Code)))
	w.raw(chunk.Code)

	w.u32(uint32(len(chunk.Constants)))

	for _, constant := range chunk.Constants {
		w.constant(constant)
	}

	w.u32(uint32(len(chunk.Names)))

	for _, name := range chunk.Names {
		w.str(name)
	}

	w.u32(uint32(chunk.NumSlots))
	w.u32(uint32(len(chunk.LocalNames)))

	for _, name := range chunk.LocalNames {
		w.str(name)
	}

	w.u32(uint32(len(chunk.Functions)))

	for _, fn := range chunk.Functions {
		w.functionTemplate(fn)
	}

	w.lineTable(chunk)
}

func (w *writer) constant(value values.Value) {
	switch val := value.(type) {
	case *values.Number:
		if val.IsInt() {
			intVal, _ := val.AsInt()
			w.u8(constTagInt)
			w.u64(uint64(intVal))
		} else {
			w.u8(constTagFloat)
			w.u64(math.Float64bits(val.AsFloat()))
		}

	case *values.String:
		w.u8(constTagString)
		w.str(val.Value)

	default:
		panic(ModenaError("cannot serialize constant of type %T", value))
	}
}

func (w *writer) functionTemplate(template *FunctionTemplate) {
	w.str(template.Name)

	w.u32(uint32(template.Arity))

	w.u32(uint32(len(template.Upvalues)))

	for _, desc := range template.Upvalues {
		fromLocal := byte(0)

		if desc.FromLocal {
			fromLocal = 1
		}

		w.u8(fromLocal)
		w.u32(uint32(desc.Index))
		w.str(desc.Name)
	}

	w.chunk(template.Chunk)
}

// lineTable writes the table that connects each instruction back to the source
// line it originated from. Neighbouring instructions almost always share a line,
// so only the points where the line or the file changes are written.

// The source text itself is never stored, so an error from a bytecode file can
// name the line but cannot print it with the ^ arrows / carets under them.
func (w *writer) lineTable(chunk *Chunk) {
	type entry struct {
		offset uint32
		line   uint32
		file   string
	}

	var entries []entry
	prevLine := -1
	prevFile := ""
	haveEntry := false

	for i := 0; i < chunk.SpanCount(); i++ {
		span := chunk.SpanAt(i)
		line := 0
		file := ""

		if span.Start != nil {
			line = span.Start.Line
			file = span.Start.File
		}

		if !haveEntry || line != prevLine || file != prevFile {
			entries = append(entries, entry{offset: uint32(i), line: uint32(line), file: file})
			prevLine = line
			prevFile = file
			haveEntry = true
		}
	}

	w.u32(uint32(len(entries)))

	for _, entry := range entries {
		w.u32(entry.offset)
		w.u32(entry.line)
		w.str(entry.file)
	}
}

type reader struct {
	buf   []byte
	pos   int
	depth int
	err   error
}

func (r *reader) fail(format string, args ...any) {
	if r.err == nil {
		r.err = ModenaError(format, args...)
	}
}

func (r *reader) take(count int) []byte {
	if r.err != nil {
		return nil
	}

	if count < 0 || r.pos+count > len(r.buf) {
		r.fail("truncated bytecode")

		return nil
	}

	taken := r.buf[r.pos : r.pos+count]
	r.pos += count

	return taken
}

func (r *reader) expectMagic() bool {
	raw := r.take(len(magic))

	return r.err == nil && bytes.Equal(raw, magic)
}

func (r *reader) u8() byte {
	raw := r.take(1)

	if r.err != nil {
		return 0
	}

	return raw[0]
}

func (r *reader) u32() uint32 {
	raw := r.take(4)

	if r.err != nil {
		return 0
	}

	return binary.LittleEndian.Uint32(raw)
}

func (r *reader) u64() uint64 {
	raw := r.take(8)

	if r.err != nil {
		return 0
	}

	return binary.LittleEndian.Uint64(raw)
}

func (r *reader) str() string {
	length := r.u32()

	raw := r.take(int(length))

	if r.err != nil {
		return ""
	}

	return string(raw)
}

func (r *reader) chunk() *Chunk {
	if r.err != nil {
		return nil
	}

	if r.depth++; r.depth > constants.LIMIT_DECODE_DEPTH {
		r.fail("bytecode nested too deep")

		return nil
	}

	defer func() { r.depth-- }()

	chunk := NewChunk()

	codeLen := int(r.u32())
	chunk.Code = append([]byte{}, r.take(codeLen)...)

	nConst := int(r.u32())

	for i := 0; i < nConst && r.err == nil; i++ {
		chunk.Constants = append(chunk.Constants, r.constant())
	}

	nName := int(r.u32())

	for i := 0; i < nName && r.err == nil; i++ {
		chunk.Names = append(chunk.Names, r.str())
	}

	chunk.NumSlots = int(r.u32())

	nLocal := int(r.u32())

	for i := 0; i < nLocal && r.err == nil; i++ {
		chunk.LocalNames = append(chunk.LocalNames, r.str())
	}

	nFunctions := int(r.u32())

	for i := 0; i < nFunctions && r.err == nil; i++ {
		chunk.Functions = append(chunk.Functions, r.functionTemplate())
	}

	r.readLineTable(chunk, codeLen)
	chunk.FinishBuilding()

	return chunk
}

func (r *reader) constant() values.Value {
	switch tag := r.u8(); tag {
	case constTagInt:
		return values.NewNumber(int64(r.u64()))

	case constTagFloat:
		num, err := values.NewNumberFromFloat(math.Float64frombits(r.u64()))

		if err != nil {
			r.fail("%v", err)

			return nil
		}

		return num

	case constTagString:
		return values.NewString(r.str())

	default:
		r.fail("unknown constant tag %d", tag)

		return nil
	}
}

func (r *reader) functionTemplate() *FunctionTemplate {
	template := &FunctionTemplate{}
	template.Name = r.str()
	template.Arity = int(r.u32())

	nUpvalue := int(r.u32())

	for i := 0; i < nUpvalue && r.err == nil; i++ {
		desc := UpvalueDesc{FromLocal: r.u8() == 1}
		desc.Index = int(r.u32())
		desc.Name = r.str()
		template.Upvalues = append(template.Upvalues, desc)
	}

	template.Chunk = r.chunk()

	return template
}

func (r *reader) readLineTable(chunk *Chunk, codeLen int) {
	type entry struct {
		offset uint32
		pos    *errors.Position
	}

	nEntries := int(r.u32())

	var entries []entry

	for i := 0; i < nEntries && r.err == nil; i++ {
		offset := r.u32()
		line := int(r.u32())
		file := r.str()
		entries = append(entries, entry{offset: offset, pos: errors.NewPosition(0, line, 0, file, "")})
	}

	if r.err != nil {
		return
	}

	current := 0

	for i := 0; i < codeLen; i++ {
		for current+1 < len(entries) && entries[current+1].offset <= uint32(i) {
			current++
		}

		var pos *errors.Position

		if len(entries) > 0 {
			pos = entries[current].pos
		}

		chunk.spanIndex = append(chunk.spanIndex, chunk.shareSpan(Span{Start: pos, End: pos}))
	}
}
