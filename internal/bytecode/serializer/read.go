/*
 *
 * Chippy - internal/bytecode/serializer/read.go
 *
 */

package serializer

import (
	"bytes"
	"chip-go/internal/bytecode"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"encoding/binary"
	"math"
)

func Deserialize(data []byte) (*bytecode.Chunk, error) {
	data = skipShebang(data)

	src := &reader{buf: data}

	if !src.expectMagic() {
		return nil, errors.ModenaError("not a bytecode file")
	}

	version := src.u32()

	if version != constants.MODENA_FORMAT_VERSION {
		return nil, errors.ModenaError("unsupported bytecode version %d, expected %d", version, constants.MODENA_FORMAT_VERSION)
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

type reader struct {
	buf   []byte
	pos   int
	depth int
	err   error
}

func (r *reader) fail(format string, args ...any) {
	if r.err == nil {
		r.err = errors.ModenaError(format, args...)
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

func (r *reader) chunk() *bytecode.Chunk {
	if r.err != nil {
		return nil
	}

	if r.depth++; r.depth > constants.LIMIT_DECODE_DEPTH {
		r.fail("bytecode nested too deep")

		return nil
	}

	defer func() { r.depth-- }()

	chunk := bytecode.NewChunk()

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

			return values.Value{}
		}

		return num

	case constTagString:
		return values.NewString(r.str())

	default:
		r.fail("unknown constant tag %d", tag)

		return values.Value{}
	}
}

func (r *reader) functionTemplate() *bytecode.FunctionTemplate {
	template := &bytecode.FunctionTemplate{}
	template.Name = r.str()
	template.Arity = int(r.u32())

	nUpvalue := int(r.u32())

	for i := 0; i < nUpvalue && r.err == nil; i++ {
		desc := bytecode.UpvalueDesc{FromLocal: r.u8() == 1}
		desc.Index = int(r.u32())
		desc.Name = r.str()
		template.Upvalues = append(template.Upvalues, desc)
	}

	template.Chunk = r.chunk()

	return template
}
