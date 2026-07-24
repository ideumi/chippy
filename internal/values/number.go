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
func NewNumber[T integerLiteral](v T) *Number {
	switch x := any(v).(type) {

	case int:
		return numberFromInt64(int64(x))

	case int8:
		return numberFromInt64(int64(x))

	case int16:
		return numberFromInt64(int64(x))

	case int32:
		return numberFromInt64(int64(x))

	case int64:
		return numberFromInt64(x)

	case uint:
		return numberFromUint64(uint64(x))

	case uint8:
		return numberFromInt64(int64(x))

	case uint16:
		return numberFromInt64(int64(x))

	case uint32:
		return numberFromInt64(int64(x))

	case uint64:
		return numberFromUint64(x)

	}

	panic("NewNumber: unreachable")
}

// NewNumberFromFloat rejects NaN and +-Infinity. Integer-valued floats in int64
// range are coerced to int storage so a later AsInt is exact.
func NewNumberFromFloat(v float64) (*Number, error) {
	if math.IsNaN(v) {
		return nil, fmt.Errorf("Numeric result is NaN")
	}

	if math.IsInf(v, 0) {
		return nil, fmt.Errorf("Numeric overflow")
	}

	return numberFromFloat64(v), nil
}

// Private helpers that bypass the NaN/Inf gate. Use only when the caller has
// already proven the input is finite.
func numberFromInt64(v int64) *Number {
	return &Number{
		BaseValue: NewBaseValue(),
		iVal:      v,
		isInt:     true,
	}
}

func numberFromUint64(v uint64) *Number {
	if v <= math.MaxInt64 {
		return numberFromInt64(int64(v))
	}

	return numberFromFloat64(float64(v))
}

func numberFromFloat64(v float64) *Number {
	n := &Number{
		BaseValue: NewBaseValue(),
		fVal:      v,
	}

	if v != math.Trunc(v) {
		return n
	}

	// float64(MaxInt64) rounds up to 2^63, so >= is the correct upper guard.
	if v < math.MinInt64 || v >= float64(math.MaxInt64) {
		return n
	}

	n.iVal = int64(v)
	n.isInt = true

	return n
}

func newBoolNumber(b bool) *Number {
	if b {
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
		return 0, errors.NewRTError(
			n.posStart, n.posEnd,
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

func addInt64(a, b int64) (int64, bool) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, true
	}

	return a + b, false
}

func subInt64(a, b int64) (int64, bool) {
	if (b < 0 && a > math.MaxInt64+b) || (b > 0 && a < math.MinInt64+b) {
		return 0, true
	}

	return a - b, false
}

func mulInt64(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, false
	}

	if a == math.MinInt64 && b == -1 {
		return 0, true
	}

	if b == math.MinInt64 && a == -1 {
		return 0, true
	}

	c := a * b

	if c/b != a {
		return 0, true
	}

	return c, false
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
		r, ov := mulInt64(result, base)

		if ov {
			return 0, true
		}

		result = r
	}

	return result, false
}

func wrapFloatResult(v float64, ctx Ctx, posStart, posEnd *errors.Position) (Value, error) {
	n, err := NewNumberFromFloat(v)

	if err != nil {
		return nil, errors.NewRTError(posStart, posEnd, err.Error())
	}

	return n.SetContext(ctx), nil
}

func (n *Number) AddedTo(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	if n.isInt && otherNum.isInt {
		sum, overflow := addInt64(n.iVal, otherNum.iVal)

		if !overflow {
			return NewNumber(sum).SetContext(n.context), nil
		}
	}

	return wrapFloatResult(n.AsFloat()+otherNum.AsFloat(), n.context, n.posStart, otherNum.posEnd)
}

func (n *Number) SubbedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	if n.isInt && otherNum.isInt {
		diff, overflow := subInt64(n.iVal, otherNum.iVal)

		if !overflow {
			return NewNumber(diff).SetContext(n.context), nil
		}
	}

	return wrapFloatResult(n.AsFloat()-otherNum.AsFloat(), n.context, n.posStart, otherNum.posEnd)
}

func (n *Number) MultedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	if n.isInt && otherNum.isInt {
		prod, overflow := mulInt64(n.iVal, otherNum.iVal)

		if !overflow {
			return NewNumber(prod).SetContext(n.context), nil
		}
	}

	return wrapFloatResult(n.AsFloat()*otherNum.AsFloat(), n.context, n.posStart, otherNum.posEnd)
}

// Exact integer divides stay in int64 so precision above 2^53 survives; inexact
// ones fall through to float.
func (n *Number) DivedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	if n.isInt && otherNum.isInt {
		if otherNum.iVal == 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
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
		return nil, errors.NewRTError(
			otherNum.posStart, otherNum.posEnd,
			"Division by zero")

	}

	return wrapFloatResult(n.AsFloat()/rhs, n.context, n.posStart, otherNum.posEnd)
}

func (n *Number) PowedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
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
		return nil, errors.NewRTError(
			n.posStart, otherNum.posEnd,
			"Negative base raised to a non-integer power")

	}

	// math.Pow(0, negative) is +Inf. Same reason.
	if base == 0 && exp < 0 {
		return nil, errors.NewRTError(
			n.posStart, otherNum.posEnd,
			"Zero raised to a negative power")

	}

	return wrapFloatResult(math.Pow(base, exp), n.context, n.posStart, otherNum.posEnd)
}

func (n *Number) ModdedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	if n.isInt && otherNum.isInt {
		if otherNum.iVal == 0 {
			return nil, errors.NewRTError(
				otherNum.posStart, otherNum.posEnd,
				"Division by zero")

		}

		if otherNum.iVal == 1 || otherNum.iVal == -1 {
			return NewNumber(0).SetContext(n.context), nil
		}

		return NewNumber(n.iVal % otherNum.iVal).SetContext(n.context), nil
	}

	rhs := otherNum.AsFloat()

	if rhs == 0 {
		return nil, errors.NewRTError(
			otherNum.posStart, otherNum.posEnd,
			"Division by zero")

	}

	return wrapFloatResult(math.Mod(n.AsFloat(), rhs), n.context, n.posStart, otherNum.posEnd)
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

func compareEqual(a, b *Number) bool {
	if a.isInt && b.isInt {
		return a.iVal == b.iVal
	}

	if a.isInt {
		return floatEqualsInt(b.fVal, a.iVal)
	}

	if b.isInt {
		return floatEqualsInt(a.fVal, b.iVal)
	}

	return a.fVal == b.fVal
}

func floatEqualsInt(f float64, i int64) bool {
	if f != math.Trunc(f) {
		return false
	}

	if f < math.MinInt64 || f >= float64(math.MaxInt64) {
		return false
	}

	return int64(f) == i
}

func compareLess(a, b *Number) bool {
	if a.isInt && b.isInt {
		return a.iVal < b.iVal
	}

	if a.isInt {
		return intLessFloat(a.iVal, b.fVal)
	}

	if b.isInt {
		return floatLessInt(a.fVal, b.iVal)
	}

	return a.fVal < b.fVal
}

func intLessFloat(i int64, f float64) bool {
	// f at or above 2^63 is greater than every int64.
	if f >= float64(math.MaxInt64) {
		return true
	}

	// f below MinInt64 is less than every int64.
	if f < math.MinInt64 {
		return false
	}

	if f == math.Trunc(f) {
		return i < int64(f)
	}

	return float64(i) < f
}

func floatLessInt(f float64, i int64) bool {
	if f >= float64(math.MaxInt64) {
		return false
	}

	if f < math.MinInt64 {
		return true
	}

	if f == math.Trunc(f) {
		return int64(f) < i
	}

	return f < float64(i)
}

func (n *Number) GetComparisonEe(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	return newBoolNumber(compareEqual(n, otherNum)).SetContext(n.context), nil
}

func (n *Number) GetComparisonNe(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	return newBoolNumber(!compareEqual(n, otherNum)).SetContext(n.context), nil
}

func (n *Number) GetComparisonLt(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	return newBoolNumber(compareLess(n, otherNum)).SetContext(n.context), nil
}

func (n *Number) GetComparisonGt(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	return newBoolNumber(compareLess(otherNum, n)).SetContext(n.context), nil
}

func (n *Number) GetComparisonLte(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	return newBoolNumber(!compareLess(otherNum, n)).SetContext(n.context), nil
}

func (n *Number) GetComparisonGte(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
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
func bitwiseInt(n *Number) (int64, error) {
	if n.isInt {
		return n.iVal, nil
	}

	if n.fVal != math.Trunc(n.fVal) {
		return 0, errors.NewRTError(
			n.posStart, n.posEnd,
			"Bitwise operation on decimal value")

	}

	if n.fVal < math.MinInt64 || n.fVal >= float64(math.MaxInt64) {
		return 0, errors.NewRTError(
			n.posStart, n.posEnd,
			"Bitwise operation on value out of integer range")

	}

	return int64(n.fVal), nil
}

func (n *Number) BAndedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	a, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	b, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	return NewNumber(a & b).SetContext(n.context), nil
}

func (n *Number) BOredBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	a, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	b, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	return NewNumber(a | b).SetContext(n.context), nil
}

func (n *Number) BNotted() (Value, error) {
	a, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	return NewNumber(^a).SetContext(n.context), nil
}

func (n *Number) BXoredBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	a, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	b, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	return NewNumber(a ^ b).SetContext(n.context), nil
}

func (n *Number) LShiftedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	a, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	b, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	if b < 0 {
		return nil, errors.NewRTError(
			otherNum.posStart, otherNum.posEnd,
			"Negative shift amount")

	}

	return NewNumber(a << uint64(b)).SetContext(n.context), nil
}

func (n *Number) RShiftedBy(other Value) (Value, error) {
	otherNum, ok := other.(*Number)

	if !ok {
		return nil, IllegalOperation(n, other)
	}

	a, err := bitwiseInt(n)

	if err != nil {
		return nil, err
	}

	b, err := bitwiseInt(otherNum)

	if err != nil {
		return nil, err
	}

	if b < 0 {
		return nil, errors.NewRTError(
			otherNum.posStart, otherNum.posEnd,
			"Negative shift amount")

	}

	return NewNumber(a >> uint64(b)).SetContext(n.context), nil
}
