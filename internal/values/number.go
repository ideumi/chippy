/*
 *
 * RR2 - internal/values/number.go
 *
 */

package values

import (
	"chip-go/internal/errors"
	"fmt"
	"math"
	"strconv"
)

// Always finite: NaN/Inf are rejected by the public constructors.
type Number struct {
	*BaseValue

	iVal  int64
	fVal  float64
	isInt bool
}

type integerLiteral interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Only for integers: float sources must use NewNumberFromFloat for the NaN/Inf
// check.
func NewNumber[T integerLiteral](val T) *Number {
	switch typed := any(val).(type) {

	case int:
		return numberFromInt64(int64(typed))

	case int8:
		return numberFromInt64(int64(typed))

	case int16:
		return numberFromInt64(int64(typed))

	case int32:
		return numberFromInt64(int64(typed))

	case int64:
		return numberFromInt64(typed)

	case uint:
		return numberFromUint64(uint64(typed))

	case uint8:
		return numberFromInt64(int64(typed))

	case uint16:
		return numberFromInt64(int64(typed))

	case uint32:
		return numberFromInt64(int64(typed))

	case uint64:
		return numberFromUint64(typed)

	}

	panic("NewNumber: unreachable")
}

// NewNumberFromFloat rejects NaN and +-Infinity. Integer-valued floats in int64
// range are coerced to int storage so a later AsInt is exact.
func NewNumberFromFloat(val float64) (*Number, error) {
	if math.IsNaN(val) {
		return nil, fmt.Errorf("Numeric result is NaN")
	}

	if math.IsInf(val, 0) {
		return nil, fmt.Errorf("Numeric overflow")
	}

	return numberFromFloat64(val), nil
}

// Private helpers that bypass the NaN/Inf gate. Use only when the caller has
// already proven the input is finite.
func numberFromInt64(val int64) *Number {
	return &Number{
		BaseValue: NewBaseValue(),
		iVal:      val,
		isInt:     true,
	}
}

func numberFromUint64(val uint64) *Number {
	if val <= math.MaxInt64 {
		return numberFromInt64(int64(val))
	}

	return numberFromFloat64(float64(val))
}

func numberFromFloat64(val float64) *Number {
	result := &Number{
		BaseValue: NewBaseValue(),
		fVal:      val,
	}

	if val != math.Trunc(val) {
		return result
	}

	// float64(MaxInt64) rounds up to 2^63, so >= is the correct upper guard.
	if val < math.MinInt64 || val >= float64(math.MaxInt64) {
		return result
	}

	result.iVal = int64(val)
	result.isInt = true

	return result
}

func newBoolNumber(flag bool) *Number {
	if flag {
		return NewNumber(1)
	}

	return NewNumber(0)
}

func (n *Number) IsInt() bool {
	return n.isInt
}

// AsInt errors when the underlying value is a float outside [MinInt64, 2^63),
// since Go's float-to-int conversion is implementation-defined there.
func (n *Number) AsInt() (int64, error) {
	if n.isInt {
		return n.iVal, nil
	}

	if n.fVal >= float64(math.MaxInt64) || n.fVal < math.MinInt64 {
		return 0, errors.NewCallError(
			"Value out of integer range")

	}

	return int64(n.fVal), nil
}

// AsFloat is infallible. Lossy for integer values with |v| > 2^53.
func (n *Number) AsFloat() float64 {
	if n.isInt {
		return float64(n.iVal)
	}

	return n.fVal
}

func (n *Number) String() string {
	if n.isInt {
		return strconv.FormatInt(n.iVal, 10)
	}

	return strconv.FormatFloat(n.fVal, 'f', -1, 64)
}

func (n *Number) SetPos(posStart, posEnd *errors.Position) Value {
	n.BaseValue.SetPos(posStart, posEnd)
	return n
}

func (n *Number) SetContext(ctx Ctx) Value {
	n.BaseValue.SetContext(ctx)
	return n
}

func (n *Number) Copy() Value {
	var copy *Number

	if n.isInt {
		copy = numberFromInt64(n.iVal)
	} else {
		copy = numberFromFloat64(n.fVal)
	}

	copy.SetPos(n.posStart, n.posEnd)
	copy.SetContext(n.context)

	return copy
}

func (n *Number) IsTrue() bool {
	if n.isInt {
		return n.iVal != 0
	}

	return n.fVal != 0
}

func addInt64(left, right int64) (int64, bool) {
	if (right > 0 && left > math.MaxInt64-right) || (right < 0 && left < math.MinInt64-right) {
		return 0, true
	}

	return left + right, false
}

func subInt64(left, right int64) (int64, bool) {
	if (right < 0 && left > math.MaxInt64+right) || (right > 0 && left < math.MinInt64+right) {
		return 0, true
	}

	return left - right, false
}

func mulInt64(left, right int64) (int64, bool) {
	if left == 0 || right == 0 {
		return 0, false
	}

	if left == math.MinInt64 && right == -1 {
		return 0, true
	}

	if right == math.MinInt64 && left == -1 {
		return 0, true
	}

	product := left * right

	if product/right != left {
		return 0, true
	}

	return product, false
}

func powInt64(base, exp int64) (int64, bool) {
	if exp < 0 {
		return 0, true
	}

	if exp == 0 {
		return 1, false
	}

	if base == 0 {
		return 0, false
	}

	if base == 1 {
		return 1, false
	}

	if base == -1 {
		if exp%2 == 0 {
			return 1, false
		}

		return -1, false
	}

	result := int64(1)

	for i := int64(0); i < exp; i++ {
		product, overflow := mulInt64(result, base)

		if overflow {
			return 0, true
		}

		result = product
	}

	return result, false
}

func wrapFloatResult(val float64, ctx Ctx) (Value, error) {
	number, err := NewNumberFromFloat(val)

	if err != nil {
		return nil, errors.NewCallError(err.Error())
	}

	return number.SetContext(ctx), nil
}

func (n *Number) AddedTo(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	if n.isInt && otherNum.isInt {
		sum, overflow := addInt64(n.iVal, otherNum.iVal)

		if !overflow {
			return NewNumber(sum).SetContext(n.context), nil
		}
	}

	return wrapFloatResult(n.AsFloat()+otherNum.AsFloat(), n.context)
}

func (n *Number) SubbedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	if n.isInt && otherNum.isInt {
		diff, overflow := subInt64(n.iVal, otherNum.iVal)

		if !overflow {
			return NewNumber(diff).SetContext(n.context), nil
		}
	}

	return wrapFloatResult(n.AsFloat()-otherNum.AsFloat(), n.context)
}

func (n *Number) MultedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	if n.isInt && otherNum.isInt {
		prod, overflow := mulInt64(n.iVal, otherNum.iVal)

		if !overflow {
			return NewNumber(prod).SetContext(n.context), nil
		}
	}

	return wrapFloatResult(n.AsFloat()*otherNum.AsFloat(), n.context)
}

// Exact integer divides stay in int64 so precision above 2^53 survives; inexact
// ones fall through to float.
func (n *Number) DivedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	if n.isInt && otherNum.isInt {
		if otherNum.iVal == 0 {
			return nil, errors.NewCallError(
				"Division by zero")

		}

		// MinInt64 / -1 overflows int64. Fall through to float.
		if !(n.iVal == math.MinInt64 && otherNum.iVal == -1) {
			if n.iVal%otherNum.iVal == 0 {
				return NewNumber(n.iVal / otherNum.iVal).SetContext(n.context), nil
			}
		}
	}

	rhs := otherNum.AsFloat()

	if rhs == 0 {
		return nil, errors.NewCallError(
			"Division by zero")

	}

	return wrapFloatResult(n.AsFloat()/rhs, n.context)
}

func (n *Number) PowedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	if n.isInt && otherNum.isInt && otherNum.iVal >= 0 {
		r, overflow := powInt64(n.iVal, otherNum.iVal)

		if !overflow {
			return NewNumber(r).SetContext(n.context), nil
		}
	}

	base := n.AsFloat()
	exp := otherNum.AsFloat()

	// math.Pow(neg, fractional) is NaN. Reject early for a precise message.
	if base < 0 && exp != math.Trunc(exp) {
		return nil, errors.NewCallError(
			"Negative base raised to a non-integer power")

	}

	// math.Pow(0, negative) is +Inf. Same reason.
	if base == 0 && exp < 0 {
		return nil, errors.NewCallError(
			"Zero raised to a negative power")

	}

	return wrapFloatResult(math.Pow(base, exp), n.context)
}

func (n *Number) ModdedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	if n.isInt && otherNum.isInt {
		if otherNum.iVal == 0 {
			return nil, errors.NewCallError(
				"Division by zero")

		}

		if otherNum.iVal == 1 || otherNum.iVal == -1 {
			return NewNumber(0).SetContext(n.context), nil
		}

		return NewNumber(n.iVal % otherNum.iVal).SetContext(n.context), nil
	}

	rhs := otherNum.AsFloat()

	if rhs == 0 {
		return nil, errors.NewCallError(
			"Division by zero")

	}

	return wrapFloatResult(math.Mod(n.AsFloat(), rhs), n.context)
}

// MinInt64 promotes to float because -MinInt64 doesn't fit in int64.
func (n *Number) Negate() *Number {
	if n.isInt {
		if n.iVal == math.MinInt64 {
			return numberFromFloat64(-float64(n.iVal)).SetContext(n.context).(*Number)
		}

		return NewNumber(-n.iVal).SetContext(n.context).(*Number)
	}

	return numberFromFloat64(-n.fVal).SetContext(n.context).(*Number)
}

func (n *Number) Negated() (Value, error) {
	return n.Negate(), nil
}

func compareEqual(left, right *Number) bool {
	if left.isInt && right.isInt {
		return left.iVal == right.iVal
	}

	if left.isInt {
		return floatEqualsInt(right.fVal, left.iVal)
	}

	if right.isInt {
		return floatEqualsInt(left.fVal, right.iVal)
	}

	return left.fVal == right.fVal
}

func floatEqualsInt(floatVal float64, intVal int64) bool {
	if floatVal != math.Trunc(floatVal) {
		return false
	}

	if floatVal < math.MinInt64 || floatVal >= float64(math.MaxInt64) {
		return false
	}

	return int64(floatVal) == intVal
}

func compareLess(left, right *Number) bool {
	if left.isInt && right.isInt {
		return left.iVal < right.iVal
	}

	if left.isInt {
		return intLessFloat(left.iVal, right.fVal)
	}

	if right.isInt {
		return floatLessInt(left.fVal, right.iVal)
	}

	return left.fVal < right.fVal
}

func intLessFloat(intVal int64, floatVal float64) bool {
	// floatVal at or above 2^63 is greater than every int64.
	if floatVal >= float64(math.MaxInt64) {
		return true
	}

	// floatVal below MinInt64 is less than every int64.
	if floatVal < math.MinInt64 {
		return false
	}

	if floatVal == math.Trunc(floatVal) {
		return intVal < int64(floatVal)
	}

	return float64(intVal) < floatVal
}

func floatLessInt(floatVal float64, intVal int64) bool {
	if floatVal >= float64(math.MaxInt64) {
		return false
	}

	if floatVal < math.MinInt64 {
		return true
	}

	if floatVal == math.Trunc(floatVal) {
		return int64(floatVal) < intVal
	}

	return floatVal < float64(intVal)
}

func (n *Number) GetComparisonEe(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	return newBoolNumber(compareEqual(n, otherNum)).SetContext(n.context), nil
}

func (n *Number) GetComparisonNe(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	return newBoolNumber(!compareEqual(n, otherNum)).SetContext(n.context), nil
}

func (n *Number) GetComparisonLt(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	return newBoolNumber(compareLess(n, otherNum)).SetContext(n.context), nil
}

func (n *Number) GetComparisonGt(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	return newBoolNumber(compareLess(otherNum, n)).SetContext(n.context), nil
}

func (n *Number) GetComparisonLte(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	return newBoolNumber(!compareLess(otherNum, n)).SetContext(n.context), nil
}

func (n *Number) GetComparisonGte(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	return newBoolNumber(!compareLess(n, otherNum)).SetContext(n.context), nil
}

func (n *Number) Notted() (Value, error) {
	return newBoolNumber(!n.IsTrue()).SetContext(n.context), nil
}

func (n *Number) XoredBy(other Value) (Value, error) {
	return newBoolNumber(n.IsTrue() != other.IsTrue()).SetContext(n.context), nil
}

// NaN/Inf cannot reach here by invariant.
func bitwiseInt(number *Number) (int64, error) {
	if number.isInt {
		return number.iVal, nil
	}

	if number.fVal != math.Trunc(number.fVal) {
		return 0, errors.NewCallError(
			"Bitwise operation on decimal value")

	}

	if number.fVal < math.MinInt64 || number.fVal >= float64(math.MaxInt64) {
		return 0, errors.NewCallError(
			"Bitwise operation on value out of integer range")

	}

	return int64(number.fVal), nil
}

func (n *Number) BAndedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	left, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	right, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	return NewNumber(left & right).SetContext(n.context), nil
}

func (n *Number) BOredBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	left, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	right, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	return NewNumber(left | right).SetContext(n.context), nil
}

func (n *Number) BNotted() (Value, error) {
	operand, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	return NewNumber(^operand).SetContext(n.context), nil
}

func (n *Number) BXoredBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	left, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	right, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	return NewNumber(left ^ right).SetContext(n.context), nil
}

func (n *Number) LShiftedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	left, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	right, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	if right < 0 {
		return nil, errors.NewCallError(
			"Negative shift amount")

	}

	return NewNumber(left << uint64(right)).SetContext(n.context), nil
}

func (n *Number) RShiftedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation()
	}

	left, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	right, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	if right < 0 {
		return nil, errors.NewCallError(
			"Negative shift amount")

	}

	return NewNumber(left >> uint64(right)).SetContext(n.context), nil
}
