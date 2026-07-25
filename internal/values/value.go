/*
 *
 * RR2 - internal/values/value.go
 *
 */

package values

import (
	"chip-go/internal/context"
	"chip-go/internal/errors"
)

type Ctx = *context.Context[Value]

type RuntimeResult struct {
	Value              Value
	Error              error
	FuncReturnValue    Value
	LoopShouldContinue bool
	LoopShouldBreak    bool
	FailArg            int
}

func NewRuntimeResult() *RuntimeResult {
	return &RuntimeResult{}
}

func (rr *RuntimeResult) reset() {
	rr.Value = nil
	rr.Error = nil
	rr.FuncReturnValue = nil
	rr.LoopShouldContinue = false
	rr.LoopShouldBreak = false
	rr.FailArg = 0
}

func (rr *RuntimeResult) Register(res *RuntimeResult) Value {
	if res.Error != nil {
		rr.Error = res.Error
	}
	rr.FuncReturnValue = res.FuncReturnValue
	rr.LoopShouldContinue = res.LoopShouldContinue
	rr.LoopShouldBreak = res.LoopShouldBreak

	return res.Value
}

func (rr *RuntimeResult) Success(value Value) *RuntimeResult {
	rr.reset()
	rr.Value = value

	return rr
}

func (rr *RuntimeResult) SuccessReturn(value Value) *RuntimeResult {
	rr.reset()
	rr.FuncReturnValue = value

	return rr
}

func (rr *RuntimeResult) SuccessContinue() *RuntimeResult {
	rr.reset()
	rr.LoopShouldContinue = true

	return rr
}

func (rr *RuntimeResult) SuccessBreak() *RuntimeResult {
	rr.reset()
	rr.LoopShouldBreak = true

	return rr
}

func (rr *RuntimeResult) Failure(err error) *RuntimeResult {
	rr.reset()
	rr.Error = err

	return rr
}

func (rr *RuntimeResult) Fail(details string) *RuntimeResult {
	return rr.Failure(errors.NewCallError(details))
}

func (rr *RuntimeResult) FailAt(arg int, details string) *RuntimeResult {
	rr.Failure(errors.NewCallError(details))
	// FailArg is set after Failure because Failure calls reset which clears it
	rr.FailArg = arg

	return rr
}

func (rr *RuntimeResult) ShouldReturn() bool {
	return rr.Error != nil || rr.FuncReturnValue != nil || rr.LoopShouldContinue || rr.LoopShouldBreak
}

type Value interface {
	String() string
	SetPos(posStart, posEnd *errors.Position) Value
	SetContext(ctx Ctx) Value
	GetPos() (*errors.Position, *errors.Position)
	GetContext() Ctx

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

	Execute(args []Value) *RuntimeResult
	Copy() Value
	IsTrue() bool
}

type BaseValue struct {
	posStart *errors.Position
	posEnd   *errors.Position
	context  Ctx
}

func NewBaseValue() *BaseValue {
	return &BaseValue{}
}

func (bv *BaseValue) SetPos(posStart, posEnd *errors.Position) Value {
	bv.posStart = posStart
	bv.posEnd = posEnd
	return bv
}

func (bv *BaseValue) SetContext(ctx Ctx) Value {
	bv.context = ctx
	return bv
}

func (bv *BaseValue) GetPos() (*errors.Position, *errors.Position) {
	return bv.posStart, bv.posEnd
}

func (bv *BaseValue) GetContext() Ctx {
	return bv.context
}

func IllegalOperation() error {
	return errors.NewCallError("Illegal operation")
}

func (bv *BaseValue) AddedTo(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) SubbedBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) MultedBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) DivedBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) PowedBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) ModdedBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) GetComparisonEe(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) GetComparisonNe(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) GetComparisonLt(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) GetComparisonGt(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) GetComparisonLte(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) GetComparisonGte(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) Notted() (Value, error) {
	return nil, errors.NewCallError("Illegal operation")
}

func (bv *BaseValue) Negated() (Value, error) {
	return nil, errors.NewCallError("Illegal operation")
}

func (bv *BaseValue) XoredBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) BAndedBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) BOredBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) BNotted() (Value, error) {
	return nil, errors.NewCallError("Illegal operation")
}

func (bv *BaseValue) BXoredBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) LShiftedBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) RShiftedBy(other Value) (Value, error) {
	return nil, IllegalOperation()
}

func (bv *BaseValue) Execute(args []Value) *RuntimeResult {
	return NewRuntimeResult().Failure(IllegalOperation())
}

func (bv *BaseValue) Copy() Value {
	panic("Copy method not implemented")
}

func (bv *BaseValue) IsTrue() bool {
	return false
}

func (bv *BaseValue) String() string {
	return "Value"
}
