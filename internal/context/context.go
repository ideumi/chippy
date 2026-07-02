/*
 *
 * RR2 - internal/context/context.go
 *
 */

package context

import "chip-go/internal/errors"

type Trace struct {
	Pos *errors.Position
}

type Context[T any] struct {
	DisplayName    string
	Parent         *Context[T]
	ParentEntryPos *errors.Position
	SymbolTable    *SymbolTable[T]
	InstanceID     int
	Trace          *Trace
}

func NewContext[T any](displayName string, parent *Context[T], parentEntryPos *errors.Position) *Context[T] {
	ctx := &Context[T]{
		DisplayName:    displayName,
		Parent:         parent,
		ParentEntryPos: parentEntryPos,
	}

	if parent != nil {
		ctx.SymbolTable = NewSymbolTable(parent.SymbolTable)
		ctx.InstanceID = parent.InstanceID
		ctx.Trace = parent.Trace
	} else {
		ctx.SymbolTable = NewSymbolTable[T](nil)
		ctx.Trace = &Trace{}
	}

	return ctx
}
