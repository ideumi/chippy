/*
 *
 * Modena - internal/bytecode/disassembler/disassembler.go
 *
 */

package disassembler

import (
	"chip-go/internal/bytecode"
	"fmt"
	"strings"
)

func Disassemble(chunk *bytecode.Chunk, name string) string {
	var out strings.Builder

	fmt.Fprintf(&out, "== %s ==\n", name)

	prevLine := -1 // The source line is noted only where it changes

	for offset := 0; offset < len(chunk.Code); {
		line := lineAt(chunk, offset)
		offset = disassembleInstruction(chunk, &out, offset, line, line != prevLine)
		prevLine = line
	}

	return out.String()
}

// DisassembleFull renders the chunk followed by every nested function it holds.
func DisassembleFull(chunk *bytecode.Chunk, name string) string {
	var out strings.Builder

	out.WriteString(Disassemble(chunk, name))
	disassembleFunctions(chunk, &out)

	return out.String()
}

func disassembleFunctions(chunk *bytecode.Chunk, out *strings.Builder) {
	for _, fn := range chunk.Functions {
		label := fn.Name

		if label == "" {
			label = "<anonymous>"
		}

		out.WriteString("\n")
		out.WriteString(Disassemble(fn.Chunk, "function "+label))
		disassembleFunctions(fn.Chunk, out)
	}
}

func lineAt(chunk *bytecode.Chunk, offset int) int {
	if offset < chunk.SpanCount() && chunk.SpanAt(offset).Start != nil {
		return chunk.SpanAt(offset).Start.DisplayLine()
	}

	return -1
}

func disassembleInstruction(chunk *bytecode.Chunk, out *strings.Builder, offset, line int, showLine bool) int {
	op := bytecode.Op(chunk.Code[offset])
	info, known := bytecode.OpTable[op]

	if !known {
		fmt.Fprintf(out, "%04d  <unknown opcode %d>\n", offset, op)

		return offset + 1
	}

	fmt.Fprintf(out, "%04d  %-16s", offset, info.Name)

	next := offset + 1

	if info.Operand != bytecode.OperandNone {
		operand := int(bytecode.ReadU32(chunk.Code, next))
		next += 4

		switch info.Operand {
		case bytecode.OperandConstant:
			if operand < len(chunk.Constants) {
				fmt.Fprintf(out, " %d (%s)", operand, chunk.Constants[operand].String())
			} else {
				fmt.Fprintf(out, " %d (<bad constant>)", operand)
			}
		case bytecode.OperandName:
			if operand < len(chunk.Names) {
				fmt.Fprintf(out, " %d (%s)", operand, chunk.Names[operand])
			} else {
				fmt.Fprintf(out, " %d (<bad name>)", operand)
			}
		case bytecode.OperandFunction:
			if operand < len(chunk.Functions) {
				fmt.Fprintf(out, " %d (%s)", operand, chunk.Functions[operand].Name)
			} else {
				fmt.Fprintf(out, " %d (<bad function>)", operand)
			}
		case bytecode.OperandJump:
			fmt.Fprintf(out, " -> %04d", operand)
		case bytecode.OperandInt:
			fmt.Fprintf(out, " %d", operand)
		}
	}

	if showLine && line > 0 {
		fmt.Fprintf(out, " ; line %d", line)
	}

	fmt.Fprintln(out)

	return next
}
