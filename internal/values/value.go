/*
 *
 * Chippy - internal/values/value.go
 *
 */

package values

import (
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
)

type Ctx = *context.Context[Value]

type Tag uint8

const (
	TagUnset Tag = 0
	TagInt   Tag = 1
	TagFloat Tag = 2
	TagText  Tag = 3
	TagList  Tag = 4
	TagMap   Tag = 5
	TagBytes Tag = 6
	TagFunc  Tag = 7
)

// Every value that is not a number is a heapObject held behind a Value's ref.
// Numbers live inline in the Value itself, so they never allocate.
type heapObject interface {
	String() string
	IsTrue() bool
	Copy() Value
	Execute(args []Value, ctx Ctx) RuntimeResult

	AddedTo(other Value) (Value, error)
	SubbedBy(other Value) (Value, error)
	MultedBy(other Value) (Value, error)
	DivedBy(other Value) (Value, error)
	PowedBy(other Value) (Value, error)
	ModdedBy(other Value) (Value, error)

	GetComparisonEe(other Value) (Value, error)
	GetComparisonNe(other Value) (Value, error)
	GetComparisonLt(other Value) (Value, error)
	GetComparisonGt(other Value) (Value, error)
	GetComparisonLte(other Value) (Value, error)
	GetComparisonGte(other Value) (Value, error)

	Notted() (Value, error)
	Negated() (Value, error)
	XoredBy(other Value) (Value, error)

	BAndedBy(other Value) (Value, error)
	BOredBy(other Value) (Value, error)
	BNotted() (Value, error)
	BXoredBy(other Value) (Value, error)
	LShiftedBy(other Value) (Value, error)
	RShiftedBy(other Value) (Value, error)
}

type Value struct {
	tag Tag

	num uint64

	ref heapObject
}

func (v Value) Tag() Tag {
	return v.tag
}

func (v Value) IsNumber() bool {
	return v.tag == TagInt || v.tag == TagFloat
}

func (v Value) IsSet() bool {
	return v.tag != TagUnset
}

func (v Value) IsUnset() bool {
	return v.tag == TagUnset
}

// FunctionObject lets the vm reach its own function type behind a value.
func (v Value) FunctionObject() any {
	if v.tag != TagFunc {
		return nil
	}

	return v.ref
}

func fromHeap(tag Tag, ref heapObject) Value {
	return Value{tag: tag, ref: ref}
}

func NewFunctionValue(fn heapObject) Value {
	return fromHeap(TagFunc, fn)
}

func IllegalOperation() error {
	return errors.NewCallError("Illegal operation")
}

func (v Value) String() string {
	switch v.tag {
	case TagUnset:
		return "null"
	case TagInt, TagFloat:
		return v.numberString()
	default:
		return v.ref.String()
	}
}

func (v Value) IsTrue() bool {
	switch v.tag {
	case TagUnset:
		return false
	case TagInt:
		return v.intVal() != 0
	case TagFloat:
		return v.floatVal() != 0
	default:
		return v.ref.IsTrue()
	}
}

func (v Value) Copy() Value {
	if v.tag == TagUnset || v.IsNumber() {
		return v
	}

	return v.ref.Copy()
}

func (v Value) Execute(args []Value, ctx Ctx) RuntimeResult {
	if v.IsNumber() {
		return NewRuntimeResult().Failure(IllegalOperation())
	}

	return v.ref.Execute(args, ctx)
}

func (v Value) AddedTo(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberAddedTo(other)
	}

	return v.ref.AddedTo(other)
}

func (v Value) SubbedBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberSubbedBy(other)
	}

	return v.ref.SubbedBy(other)
}

func (v Value) MultedBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberMultedBy(other)
	}

	return v.ref.MultedBy(other)
}

func (v Value) DivedBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberDivedBy(other)
	}

	return v.ref.DivedBy(other)
}

func (v Value) PowedBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberPowedBy(other)
	}

	return v.ref.PowedBy(other)
}

func (v Value) ModdedBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberModdedBy(other)
	}

	return v.ref.ModdedBy(other)
}

func (v Value) GetComparisonEe(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberComparison(other, comparisonEe)
	}

	return v.ref.GetComparisonEe(other)
}

func (v Value) GetComparisonNe(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberComparison(other, comparisonNe)
	}

	return v.ref.GetComparisonNe(other)
}

func (v Value) GetComparisonLt(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberComparison(other, comparisonLt)
	}

	return v.ref.GetComparisonLt(other)
}

func (v Value) GetComparisonGt(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberComparison(other, comparisonGt)
	}

	return v.ref.GetComparisonGt(other)
}

func (v Value) GetComparisonLte(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberComparison(other, comparisonLte)
	}

	return v.ref.GetComparisonLte(other)
}

func (v Value) GetComparisonGte(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberComparison(other, comparisonGte)
	}

	return v.ref.GetComparisonGte(other)
}

func (v Value) Notted() (Value, error) {
	if v.IsNumber() {
		return Bool(!v.IsTrue()), nil
	}

	return v.ref.Notted()
}

func (v Value) Negated() (Value, error) {
	if v.IsNumber() {
		return v.numberNegated(), nil
	}

	return v.ref.Negated()
}

func (v Value) XoredBy(other Value) (Value, error) {
	if v.IsNumber() {
		return Bool(v.IsTrue() != other.IsTrue()), nil
	}

	return v.ref.XoredBy(other)
}

func (v Value) BAndedBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberBitwise(other, bitwiseAnd)
	}

	return v.ref.BAndedBy(other)
}

func (v Value) BOredBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberBitwise(other, bitwiseOr)
	}

	return v.ref.BOredBy(other)
}

func (v Value) BNotted() (Value, error) {
	if v.IsNumber() {
		return v.numberBNotted()
	}

	return v.ref.BNotted()
}

func (v Value) BXoredBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberBitwise(other, bitwiseXor)
	}

	return v.ref.BXoredBy(other)
}

func (v Value) LShiftedBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberShift(other, shiftLeft)
	}

	return v.ref.LShiftedBy(other)
}

func (v Value) RShiftedBy(other Value) (Value, error) {
	if v.IsNumber() {
		return v.numberShift(other, shiftRight)
	}

	return v.ref.RShiftedBy(other)
}

// OperatorDefaults gives every heap type a default of "illegal operation" for
// each operator, so a type only implements the ones it actually supports.
type OperatorDefaults struct{}

func (OperatorDefaults) String() string {
	return "Value"
}

func (OperatorDefaults) IsTrue() bool {
	return false
}

func (OperatorDefaults) Copy() Value {
	errors.ModenaPanic("Copy not implemented")

	return Value{}
}

func (OperatorDefaults) Execute(args []Value, ctx Ctx) RuntimeResult {
	return NewRuntimeResult().Failure(IllegalOperation())
}

func (OperatorDefaults) AddedTo(other Value) (Value, error)  { return Value{}, IllegalOperation() }
func (OperatorDefaults) SubbedBy(other Value) (Value, error) { return Value{}, IllegalOperation() }
func (OperatorDefaults) MultedBy(other Value) (Value, error) { return Value{}, IllegalOperation() }
func (OperatorDefaults) DivedBy(other Value) (Value, error)  { return Value{}, IllegalOperation() }
func (OperatorDefaults) PowedBy(other Value) (Value, error)  { return Value{}, IllegalOperation() }
func (OperatorDefaults) ModdedBy(other Value) (Value, error) { return Value{}, IllegalOperation() }

func (OperatorDefaults) GetComparisonEe(other Value) (Value, error) {
	return Value{}, IllegalOperation()
}
func (OperatorDefaults) GetComparisonNe(other Value) (Value, error) {
	return Value{}, IllegalOperation()
}
func (OperatorDefaults) GetComparisonLt(other Value) (Value, error) {
	return Value{}, IllegalOperation()
}
func (OperatorDefaults) GetComparisonGt(other Value) (Value, error) {
	return Value{}, IllegalOperation()
}
func (OperatorDefaults) GetComparisonLte(other Value) (Value, error) {
	return Value{}, IllegalOperation()
}
func (OperatorDefaults) GetComparisonGte(other Value) (Value, error) {
	return Value{}, IllegalOperation()
}

func (OperatorDefaults) Notted() (Value, error)             { return Value{}, IllegalOperation() }
func (OperatorDefaults) Negated() (Value, error)            { return Value{}, IllegalOperation() }
func (OperatorDefaults) XoredBy(other Value) (Value, error) { return Value{}, IllegalOperation() }

func (OperatorDefaults) BAndedBy(other Value) (Value, error)   { return Value{}, IllegalOperation() }
func (OperatorDefaults) BOredBy(other Value) (Value, error)    { return Value{}, IllegalOperation() }
func (OperatorDefaults) BNotted() (Value, error)               { return Value{}, IllegalOperation() }
func (OperatorDefaults) BXoredBy(other Value) (Value, error)   { return Value{}, IllegalOperation() }
func (OperatorDefaults) LShiftedBy(other Value) (Value, error) { return Value{}, IllegalOperation() }
func (OperatorDefaults) RShiftedBy(other Value) (Value, error) { return Value{}, IllegalOperation() }

// The As* helpers unwrap a value's heap object, and report false for a number or
// for the wrong kind.
func AsString(value Value) (*String, bool) {
	object, ok := value.ref.(*String)
	return object, ok
}

func AsList(value Value) (*List, bool) {
	object, ok := value.ref.(*List)
	return object, ok
}

func AsMap(value Value) (*Map, bool) {
	object, ok := value.ref.(*Map)
	return object, ok
}

func AsBytes(value Value) (*Bytes, bool) {
	object, ok := value.ref.(*Bytes)
	return object, ok
}

func AsBuiltIn(value Value) (*BuiltInFunction, bool) {
	object, ok := value.ref.(*BuiltInFunction)
	return object, ok
}

func AsCallable(value Value) (Callable, bool) {
	object, ok := value.ref.(Callable)
	return object, ok
}

func AsBoundaryClosure(value Value) (BoundaryClosure, bool) {
	object, ok := value.ref.(BoundaryClosure)
	return object, ok
}

func AsInts(left, right Value) (int64, int64, bool) {
	return int64(left.num), int64(right.num), left.tag == TagInt && right.tag == TagInt
}

func enterWalk(val Value, seen map[Value]bool, depth int) (func(), error) {
	if depth > constants.LIMIT_VALUE_NESTING_DEPTH {
		return nil, errors.NewCallError(constants.E_VALUE_TOO_DEEP)
	}

	if seen[val] {
		return nil, errors.NewCallError(constants.E_CYCLIC_VALUE)
	}

	seen[val] = true

	return func() { delete(seen, val) }, nil
}

// EnterWalk is enterWalk for a walk with nowhere to return an error, such as Copy
// and sort.
func EnterWalk(val Value, seen map[Value]bool, depth int) func() {
	leave, err := enterWalk(val, seen, depth)

	if err != nil {
		panic(err)
	}

	return leave
}

func walkString(val Value, seen map[Value]bool, depth int) string {
	switch val.tag {
	case TagUnset:
		return "null"
	case TagList:
		return val.ref.(*List).stringWalk(seen, depth)
	case TagMap:
		return val.ref.(*Map).stringWalk(seen, depth)
	default:
		return val.String()
	}
}

func walkCopy(val Value, seen map[Value]bool, depth int) Value {
	switch val.tag {
	case TagList:
		return val.ref.(*List).copyWalk(seen, depth)
	case TagMap:
		return val.ref.(*Map).copyWalk(seen, depth)
	default:
		return val.Copy()
	}
}

func walkEqual(left, right Value, seen map[Value]bool, depth int) (Value, error) {
	if left.tag == TagList && right.tag == TagList {
		return left.ref.(*List).eqWalk(right, seen, depth)
	}

	if left.tag == TagMap && right.tag == TagMap {
		return left.ref.(*Map).eqWalk(right, seen, depth)
	}

	return left.GetComparisonEe(right)
}
