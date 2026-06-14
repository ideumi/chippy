/*
 *
 * RR2 - internal/values/bytes.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"fmt"
	"strings"
)

type Bytes struct {
	*BaseValue
	Data []byte
}

func NewBytes(data []byte) *Bytes {
	return &Bytes{
		BaseValue: NewBaseValue(),
		Data:      data,
	}
}

func (b *Bytes) String() string {
	if len(b.Data) == 0 {
		return "b[]"
	}

	var elements []string

	for _, byteVal := range b.Data {
		elements = append(elements, fmt.Sprintf("%d", byteVal))
	}

	return "b[" + strings.Join(elements, ", ") + "]"
}

func (b *Bytes) SetPos(posStart, posEnd *errors.Position) Value {
	b.BaseValue.SetPos(posStart, posEnd)

	return b
}

func (b *Bytes) SetContext(ctx Ctx) Value {
	b.BaseValue.SetContext(ctx)

	return b
}

func (b *Bytes) Copy() Value {
	dataCopy := make([]byte, len(b.Data))
	copy(dataCopy, b.Data)

	newBytes := NewBytes(dataCopy)
	newBytes.SetPos(b.posStart, b.posEnd)
	newBytes.SetContext(b.context)

	return newBytes
}

func (b *Bytes) IsTrue() bool {
	return len(b.Data) > 0
}

func (b *Bytes) Length() int {
	return len(b.Data)
}

func (b *Bytes) GetElement(index int) *Number {
	if index < 1 || index > len(b.Data) {
		return nil
	}

	return NewNumber(b.Data[index-1]).SetContext(b.context).(*Number)
}

func (b *Bytes) AppendByte(value int) *Bytes {
	if value < 0 || value > 255 {
		return b
	}

	b.Data = append(b.Data, byte(value))

	return b
}

func (b *Bytes) ToList() *List {
	elements := make([]Value, len(b.Data))

	for i, byteVal := range b.Data {
		elements[i] = NewNumber(byteVal).SetContext(b.context)
	}

	return NewList(elements).SetPos(b.posStart, b.posEnd).SetContext(b.context).(*List)
}

func (b *Bytes) GetComparisonEe(other Value) (Value, error) {
	if otherBytes, ok := other.(*Bytes); ok {
		result := constants.NUM_FAL

		if len(b.Data) == len(otherBytes.Data) {
			equal := true

			for i, byte1 := range b.Data {
				if byte1 != otherBytes.Data[i] {
					equal = false
					break
				}
			}

			if equal {
				result = constants.NUM_TRU
			}
		}

		return NewNumber(result).SetContext(b.context), nil
	}

	return nil, IllegalOperation(b, other)
}

func (b *Bytes) GetComparisonNe(other Value) (Value, error) {
	if otherBytes, ok := other.(*Bytes); ok {
		result := constants.NUM_TRU

		if len(b.Data) == len(otherBytes.Data) {
			equal := true

			for i, byte1 := range b.Data {
				if byte1 != otherBytes.Data[i] {
					equal = false
					break
				}
			}

			if equal {
				result = constants.NUM_FAL
			}
		}

		return NewNumber(result).SetContext(b.context), nil
	}

	return nil, IllegalOperation(b, other)
}

func (b *Bytes) Notted() (Value, error) {
	result := constants.NUM_TRU

	if b.IsTrue() {
		result = constants.NUM_FAL
	}

	return NewNumber(result).SetContext(b.context), nil
}

func (b *Bytes) XoredBy(other Value) (Value, error) {
	result := constants.NUM_FAL

	if b.IsTrue() != other.IsTrue() {
		result = constants.NUM_TRU
	}

	return NewNumber(result).SetContext(b.context), nil
}

func (b *Bytes) MultedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {

		if !otherNum.IsInt() {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Repeat count must be an integer")

		}

		if otherNum.iVal < 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Cannot repeat bytes negative times")

		}

		repeatCount := int(otherNum.iVal)
		newData := make([]byte, len(b.Data)*repeatCount)

		for i := 0; i < repeatCount; i++ {
			copy(newData[i*len(b.Data):], b.Data)
		}

		return NewBytes(newData).SetContext(b.context), nil
	}

	return nil, IllegalOperation(b, other)
}

func (b *Bytes) SubbedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {

		if !otherNum.IsInt() {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Index must be an integer")

		}

		index := int(otherNum.iVal)

		if index < 1 || index > len(b.Data) {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Byte at this index could not be removed from bytes because index is out of bounds")

		}

		newData := make([]byte, len(b.Data)-1)

		copy(newData[:index-1], b.Data[:index-1])
		copy(newData[index-1:], b.Data[index:])

		return NewBytes(newData).SetContext(b.context), nil
	}

	return nil, IllegalOperation(b, other)
}

func (b *Bytes) AddedTo(other Value) (Value, error) {

	if otherBytes, ok := other.(*Bytes); ok {
		newData := make([]byte, len(b.Data)+len(otherBytes.Data))

		copy(newData, b.Data)
		copy(newData[len(b.Data):], otherBytes.Data)

		return NewBytes(newData).SetContext(b.context), nil
	}

	return nil, IllegalOperation(b, other)
}
