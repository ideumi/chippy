/*
 *
 * Chippy - internal/context/context.go
 *
 */

package context

import "chip-go/internal/errors"

type Trace struct {
	Pos *errors.Position
}

type Storable interface {
	IsSet() bool
}

type Context[T Storable] struct {
	DisplayName    string
	Parent         *Context[T]
	ParentEntryPos *errors.Position
	Globals        *Globals[T]
	InstanceID     int
	Trace          *Trace
}

func NewContext[T Storable](displayName string, parent *Context[T], parentEntryPos *errors.Position) *Context[T] {
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
