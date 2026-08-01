/*
 *
 * Chippy - internal/values/bytes.go
 *
 */

package values

import (
	"chip-go/internal/errors"
	"fmt"
	"strings"
)

type Bytes struct {
	OperatorDefaults
	Data []byte
}

func newBytesObject(data []byte) *Bytes {
	return &Bytes{Data: data}
}

func NewBytes(data []byte) Value {
	return fromHeap(TagBytes, newBytesObject(data))
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

func (b *Bytes) Copy() Value {
	dataCopy := make([]byte, len(b.Data))
	copy(dataCopy, b.Data)

	return NewBytes(dataCopy)
}

func (b *Bytes) IsTrue() bool {
	return len(b.Data) > 0
}

func (b *Bytes) AppendByte(value int) *Bytes {
	b.Data = append(b.Data, byte(value))

	return b
}

func (b *Bytes) bytesEqual(other Value) (isBytes, equal bool) {
	otherBytes, ok := AsBytes(other)

	if !ok {
		return false, false
	}

	if len(b.Data) != len(otherBytes.Data) {
		return true, false
	}

	for i, byteVal := range b.Data {
		if byteVal != otherBytes.Data[i] {
			return true, false
		}
	}

	return true, true
}

func (b *Bytes) GetComparisonEe(other Value) (Value, error) {
	isBytes, equal := b.bytesEqual(other)

	if !isBytes {
		return Value{}, IllegalOperation()
	}

	return Bool(equal), nil
}

func (b *Bytes) GetComparisonNe(other Value) (Value, error) {
	isBytes, equal := b.bytesEqual(other)

	if !isBytes {
		return Value{}, IllegalOperation()
	}

	return Bool(!equal), nil
}

func (b *Bytes) Notted() (Value, error) {
	return Bool(!b.IsTrue()), nil
}

func (b *Bytes) XoredBy(other Value) (Value, error) {
	return Bool(b.IsTrue() != other.IsTrue()), nil
}

func (b *Bytes) MultedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if !other.IsInt() {
		return Value{}, errors.NewCallError("Repeat count must be an integer")
	}

	count, _ := other.AsInt()

	if count < 0 {
		return Value{}, errors.NewCallError("Cannot repeat bytes negative times")
	}

	repeatCount := int(count)
	newData := make([]byte, len(b.Data)*repeatCount)

	for i := 0; i < repeatCount; i++ {
		copy(newData[i*len(b.Data):], b.Data)
	}

	return NewBytes(newData), nil
}

func (b *Bytes) SubbedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if !other.IsInt() {
		return Value{}, errors.NewCallError("Index must be an integer")
	}

	position, _ := other.AsInt()
	index := int(position)

	if index < 1 || index > len(b.Data) {
		return Value{}, errors.NewCallError(
			"Byte at this index could not be removed from bytes because index is out of bounds")
	}

	newData := make([]byte, len(b.Data)-1)

	copy(newData[:index-1], b.Data[:index-1])
	copy(newData[index-1:], b.Data[index:])

	return NewBytes(newData), nil
}

func (b *Bytes) AddedTo(other Value) (Value, error) {
	otherBytes, ok := AsBytes(other)

	if !ok {
		return Value{}, IllegalOperation()
	}

	newData := make([]byte, len(b.Data)+len(otherBytes.Data))

	copy(newData, b.Data)
	copy(newData[len(b.Data):], otherBytes.Data)

	return NewBytes(newData), nil
}
