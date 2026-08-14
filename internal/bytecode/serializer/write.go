/*
 *
 * Chippy - internal/bytecode/serializer/write.go
 *
 */

package serializer

import (
	"chip-go/internal/bytecode"
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

func Serialize(chunk *bytecode.Chunk, addShebang bool) []byte {
	out := &writer{}

	if addShebang {
		out.raw([]byte(constants.CHIPPY_SHEBANG + "\n"))
	}

	out.raw(magic)
	out.u32(constants.MODENA_FORMAT_VERSION)
	out.chunk(chunk)

	return out.buf
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

func (w *writer) chunk(chunk *bytecode.Chunk) {
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
}

func (w *writer) constant(value values.Value) {
	if value.IsNumber() {
		if value.IsInt() {
			intVal, _ := value.AsInt()
			w.u8(constTagInt)
			w.u64(uint64(intVal))
		} else {
			w.u8(constTagFloat)
			w.u64(math.Float64bits(value.AsFloat()))
		}

		return
	}

	if str, ok := values.AsString(value); ok {
		w.u8(constTagString)
		w.str(str.Value)

		return
	}

	errors.ModenaPanic("cannot serialize constant of tag %d", value.Tag())
}

func (w *writer) functionTemplate(template *bytecode.FunctionTemplate) {
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
