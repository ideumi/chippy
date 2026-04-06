/*
 *
 * RR2 - internal/values/number.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"math"
	"strconv"
)

type Number struct {
	*BaseValue
	Value float64
}

func NewNumber(value float64) *Number {
	return &Number{
		BaseValue: NewBaseValue(),
		Value:     value,
	}
}

func (n *Number) String() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

func (n *Number) SetPos(posStart, posEnd *errors.Position) Value {
	n.BaseValue.SetPos(posStart, posEnd)
	return n
}

func (n *Number) SetContext(context interface{}) Value {
	n.BaseValue.SetContext(context)
	return n
}

func (n *Number) Copy() Value {
	copy := NewNumber(n.Value)
	copy.SetPos(n.posStart, n.posEnd)
	copy.SetContext(n.context)

	return copy
}

func (n *Number) IsTrue() bool {
	return n.Value != 0
}

func (n *Number) AddedTo(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		result := NewNumber(n.Value + otherNum.Value)
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) SubbedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		result := NewNumber(n.Value - otherNum.Value)
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) MultedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		result := NewNumber(n.Value * otherNum.Value)
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) DivedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		if otherNum.Value == 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Division by zero",
				n.context,
			)
		}
		result := NewNumber(n.Value / otherNum.Value)
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) PowedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		result := NewNumber(math.Pow(n.Value, otherNum.Value))
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) ModdedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		if otherNum.Value == 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Division by zero",
				n.context,
			)
		}
		result := NewNumber(math.Mod(n.Value, otherNum.Value))
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) GetComparisonEe(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		var result float64

		if n.Value == otherNum.Value {
			result = constants.NUM_TRU
		} else {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(n.context), nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) GetComparisonNe(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		var result float64

		if n.Value != otherNum.Value {
			result = constants.NUM_TRU
		} else {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(n.context), nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) GetComparisonLt(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		var result float64

		if n.Value < otherNum.Value {
			result = constants.NUM_TRU
		} else {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(n.context), nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) GetComparisonGt(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		var result float64

		if n.Value > otherNum.Value {
			result = constants.NUM_TRU
		} else {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(n.context), nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) GetComparisonLte(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		var result float64

		if n.Value <= otherNum.Value {
			result = constants.NUM_TRU
		} else {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(n.context), nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) GetComparisonGte(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		var result float64

		if n.Value >= otherNum.Value {
			result = constants.NUM_TRU
		} else {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(n.context), nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) AndedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		var result float64

		if n.IsTrue() && otherNum.IsTrue() {
			result = constants.NUM_TRU
		} else {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(n.context), nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) OredBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		var result float64

		if n.IsTrue() || otherNum.IsTrue() {
			result = constants.NUM_TRU
		} else {
			result = constants.NUM_FAL
		}

		return NewNumber(result).SetContext(n.context), nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) Notted() (Value, error) {
	var result float64

	if n.IsTrue() {
		result = constants.NUM_FAL
	} else {
		result = constants.NUM_TRU
	}

	return NewNumber(result).SetContext(n.context), nil
}

func (n *Number) XoredBy(other Value) (Value, error) {
	var result float64

	if n.IsTrue() != other.IsTrue() {
		result = constants.NUM_TRU
	} else {
		result = constants.NUM_FAL
	}

	return NewNumber(result).SetContext(n.context), nil
}

func (n *Number) BAndedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		result := NewNumber(float64(int64(n.Value) & int64(otherNum.Value)))
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) BOredBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		result := NewNumber(float64(int64(n.Value) | int64(otherNum.Value)))
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) BNotted() (Value, error) {
	result := NewNumber(float64(^int64(n.Value)))
	result.SetContext(n.context)

	return result, nil
}

func (n *Number) BXoredBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		result := NewNumber(float64(int64(n.Value) ^ int64(otherNum.Value)))
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) LShiftedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		if otherNum.Value < 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Negative shift amount",
				n.context,
			)
		}

		result := NewNumber(float64(int64(n.Value) << uint64(otherNum.Value)))
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}

func (n *Number) RShiftedBy(other Value) (Value, error) {
	if otherNum, ok := other.(*Number); ok {
		if otherNum.Value < 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Negative shift amount",
				n.context,
			)
		}

		result := NewNumber(float64(int64(n.Value) >> uint64(otherNum.Value)))
		result.SetContext(n.context)

		return result, nil
	}

	return nil, IllegalOperation(n, other)
}
