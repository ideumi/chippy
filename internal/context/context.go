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

type GlobalStore[T any] interface {
	GetByName(name string) T
	SetByName(name string, value T)
	ForEach(fn func(name string, value T))
}

type Context[T any] struct {
	DisplayName    string
	Parent         *Context[T]
	ParentEntryPos *errors.Position
	Globals        GlobalStore[T]
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
		ctx.InstanceID = parent.InstanceID
		ctx.Trace = parent.Trace
	} else {
		ctx.Trace = &Trace{}
	}

	return ctx
}
