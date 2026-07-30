/*
 *
 * Modena - internal/values/number.go
 *
 */

package values

import (
	"chip-go/internal/errors"
	"fmt"
	"math"
	"strconv"
)

type integerLiteral interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func (v Value) intVal() int64 {
	return int64(v.num)
}

func (v Value) floatVal() float64 {
	return math.Float64frombits(v.num)
}

func Int(val int64) Value {
	return Value{tag: TagInt, num: uint64(val)}
}

func Bool(flag bool) Value {
	if flag {
		return Int(1)
	}

	return Int(0)
}

// Only for integers, floats must use NewNumberFromFloat for the NaN/Inf check.
func NewNumber[T integerLiteral](val T) Value {
	switch typed := any(val).(type) {
	case int:
		return Int(int64(typed))
	case int8:
		return Int(int64(typed))
	case int16:
		return Int(int64(typed))
	case int32:
		return Int(int64(typed))
	case int64:
		return Int(typed)
	case uint:
		return numberFromUint64(uint64(typed))
	case uint8:
		return Int(int64(typed))
	case uint16:
		return Int(int64(typed))
	case uint32:
		return Int(int64(typed))
	case uint64:
		return numberFromUint64(typed)
	}

	panic("NewNumber: unreachable")
}

func numberFromUint64(val uint64) Value {
	if val <= math.MaxInt64 {
		return Int(int64(val))
	}

	return numberFromFloat(float64(val))
}

func numberFromFloat(val float64) Value {
	if val == math.Trunc(val) && val >= math.MinInt64 && val < float64(math.MaxInt64) {
		return Int(int64(val))
	}

	return Value{tag: TagFloat, num: math.Float64bits(val)}
}

func NewNumberFromFloat(val float64) (Value, error) {
	if math.IsNaN(val) {
		return Value{}, fmt.Errorf("Numeric result is NaN")
	}

	if math.IsInf(val, 0) {
		return Value{}, fmt.Errorf("Numeric overflow")
	}

	return numberFromFloat(val), nil
}

func numberFromFloatChecked(val float64) (Value, error) {
	result, err := NewNumberFromFloat(val)

	if err != nil {
		return Value{}, errors.NewCallError(err.Error())
	}

	return result, nil
}

func (v Value) IsInt() bool {
	return v.tag == TagInt
}

// go build -gcflags="-m=2" ./internal/values/ 2>&1 | grep 'inline Value.AsInt'
func (v Value) AsInt() (int64, error) {
	if v.tag == TagInt {
		return int64(v.num), nil
	}

	return v.asIntSlow()
}

//go:noinline
func (v Value) asIntSlow() (int64, error) {
	if v.tag != TagFloat {
		return 0, errors.NewCallError("Value is not a number")
	}

	if v.floatVal() >= float64(math.MaxInt64) || v.floatVal() < math.MinInt64 {
		return 0, errors.NewCallError("Value out of integer range")
	}

	return int64(v.floatVal()), nil
}

func (v Value) AsFloat() float64 {
	if v.tag == TagInt {
		return float64(v.intVal())
	}

	return v.floatVal()
}

func (v Value) numberString() string {
	if v.tag == TagInt {
		return strconv.FormatInt(v.intVal(), 10)
	}

	return strconv.FormatFloat(v.floatVal(), 'f', -1, 64)
}

func AddInt64(left, right int64) (int64, bool) {
	if (right > 0 && left > math.MaxInt64-right) || (right < 0 && left < math.MinInt64-right) {
		return 0, true
	}

	return left + right, false
}

func SubInt64(left, right int64) (int64, bool) {
	if (right < 0 && left > math.MaxInt64+right) || (right > 0 && left < math.MinInt64+right) {
		return 0, true
	}

	return left - right, false
}

func MulInt64(left, right int64) (int64, bool) {
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
		product, overflow := MulInt64(result, base)

		if overflow {
			return 0, true
		}

		result = product
	}

	return result, false
}

func (v Value) numberAddedTo(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if v.tag == TagInt && other.tag == TagInt {
		sum, overflow := AddInt64(v.intVal(), other.intVal())

		if !overflow {
			return Int(sum), nil
		}
	}

	return numberFromFloatChecked(v.AsFloat() + other.AsFloat())
}

func (v Value) numberSubbedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if v.tag == TagInt && other.tag == TagInt {
		diff, overflow := SubInt64(v.intVal(), other.intVal())

		if !overflow {
			return Int(diff), nil
		}
	}

	return numberFromFloatChecked(v.AsFloat() - other.AsFloat())
}

func (v Value) numberMultedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if v.tag == TagInt && other.tag == TagInt {
		prod, overflow := MulInt64(v.intVal(), other.intVal())

		if !overflow {
			return Int(prod), nil
		}
	}

	return numberFromFloatChecked(v.AsFloat() * other.AsFloat())
}

func (v Value) numberDivedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if v.tag == TagInt && other.tag == TagInt {
		if other.intVal() == 0 {
			return Value{}, errors.NewCallError("Division by zero")
		}

		if !(v.intVal() == math.MinInt64 && other.intVal() == -1) {
			if v.intVal()%other.intVal() == 0 {
				return Int(v.intVal() / other.intVal()), nil
			}
		}
	}

	rhs := other.AsFloat()

	if rhs == 0 {
		return Value{}, errors.NewCallError("Division by zero")
	}

	return numberFromFloatChecked(v.AsFloat() / rhs)
}

func (v Value) numberPowedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if v.tag == TagInt && other.tag == TagInt && other.intVal() >= 0 {
		result, overflow := powInt64(v.intVal(), other.intVal())

		if !overflow {
			return Int(result), nil
		}
	}

	base := v.AsFloat()
	exp := other.AsFloat()

	// math.Pow(neg, fractional) is NaN
	if base < 0 && exp != math.Trunc(exp) {
		return Value{}, errors.NewCallError("Negative base raised to a non-integer power")
	}

	// math.Pow(0, negative) is +Inf
	if base == 0 && exp < 0 {
		return Value{}, errors.NewCallError("Zero raised to a negative power")
	}

	return numberFromFloatChecked(math.Pow(base, exp))
}

func (v Value) numberModdedBy(other Value) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	if v.tag == TagInt && other.tag == TagInt {
		if other.intVal() == 0 {
			return Value{}, errors.NewCallError("Division by zero")
		}

		// Every number divides by 1 with nothing left over, and -1 is
		// here because MinInt64 % -1 overflows and panics in Go.
		if other.intVal() == 1 || other.intVal() == -1 {
			return Int(0), nil
		}

		return Int(v.intVal() % other.intVal()), nil
	}

	rhs := other.AsFloat()

	if rhs == 0 {
		return Value{}, errors.NewCallError("Division by zero")
	}

	return numberFromFloatChecked(math.Mod(v.AsFloat(), rhs))
}

func (v Value) numberNegated() Value {
	if v.tag == TagInt {
		if v.intVal() == math.MinInt64 {
			return numberFromFloat(-float64(v.intVal()))
		}

		return Int(-v.intVal())
	}

	return numberFromFloat(-v.floatVal())
}

type comparisonKind int

const (
	comparisonEe  comparisonKind = 0
	comparisonNe  comparisonKind = 1
	comparisonLt  comparisonKind = 2
	comparisonGt  comparisonKind = 3
	comparisonLte comparisonKind = 4
	comparisonGte comparisonKind = 5
)

func (v Value) numberComparison(other Value, kind comparisonKind) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	var result bool

	switch kind {
	case comparisonEe:
		result = numberEqual(v, other)
	case comparisonNe:
		result = !numberEqual(v, other)
	case comparisonLt:
		result = numberLess(v, other)
	case comparisonGt:
		result = numberLess(other, v)
	case comparisonLte:
		result = !numberLess(other, v)
	case comparisonGte:
		result = !numberLess(v, other)
	}

	return Bool(result), nil
}

func numberEqual(left, right Value) bool {
	if left.tag == TagInt && right.tag == TagInt {
		return left.intVal() == right.intVal()
	}

	if left.tag == TagInt {
		return floatEqualsInt(right.floatVal(), left.intVal())
	}

	if right.tag == TagInt {
		return floatEqualsInt(left.floatVal(), right.intVal())
	}

	return left.floatVal() == right.floatVal()
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

func numberLess(left, right Value) bool {
	if left.tag == TagInt && right.tag == TagInt {
		return left.intVal() < right.intVal()
	}

	if left.tag == TagInt {
		return intLessFloat(left.intVal(), right.floatVal())
	}

	if right.tag == TagInt {
		return floatLessInt(left.floatVal(), right.intVal())
	}

	return left.floatVal() < right.floatVal()
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

// NaN/Inf cannot reach here by invariant.
func (v Value) bitwiseInt() (int64, error) {
	if v.tag == TagInt {
		return v.intVal(), nil
	}

	if v.floatVal() != math.Trunc(v.floatVal()) {
		return 0, errors.NewCallError("Bitwise operation on decimal value")
	}

	if v.floatVal() < math.MinInt64 || v.floatVal() >= float64(math.MaxInt64) {
		return 0, errors.NewCallError("Bitwise operation on value out of integer range")
	}

	return int64(v.floatVal()), nil
}

type bitwiseKind int

const (
	bitwiseAnd bitwiseKind = 0
	bitwiseOr  bitwiseKind = 1
	bitwiseXor bitwiseKind = 2
)

func (v Value) numberBitwise(other Value, kind bitwiseKind) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	left, err := v.bitwiseInt()

	if err != nil {
		return Value{}, err
	}

	right, err := other.bitwiseInt()

	if err != nil {
		return Value{}, err
	}

	switch kind {
	case bitwiseAnd:
		return Int(left & right), nil
	case bitwiseOr:
		return Int(left | right), nil
	default:
		return Int(left ^ right), nil
	}
}

func (v Value) numberBNotted() (Value, error) {
	operand, err := v.bitwiseInt()

	if err != nil {
		return Value{}, err
	}

	return Int(^operand), nil
}

type shiftKind int

const (
	shiftLeft  shiftKind = 0
	shiftRight shiftKind = 1
)

func (v Value) numberShift(other Value, kind shiftKind) (Value, error) {
	if !other.IsNumber() {
		return Value{}, IllegalOperation()
	}

	left, err := v.bitwiseInt()

	if err != nil {
		return Value{}, err
	}

	right, err := other.bitwiseInt()

	if err != nil {
		return Value{}, err
	}

	if right < 0 {
		return Value{}, errors.NewCallError("Negative shift amount")
	}

	if kind == shiftLeft {
		return Int(left << uint64(right)), nil
	}

	return Int(left >> uint64(right)), nil
}
