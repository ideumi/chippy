/*
 *
 * RR2 - internal/values/callable.go
 *
 */

package values

type Callable interface {
	Value
	ArgCount() int
	CallableName() string
}

type BoundaryClosure interface {
	Value
	TransferCells() []*Value
	SetTransferCells(cells []*Value)
	RebindGlobals(globals Ctx)
}
