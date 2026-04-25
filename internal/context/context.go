/*
 *
 * RR2 - internal/context/context.go
 *
 */

package context

import "chip-go/internal/errors"

type Context struct {
	DisplayName    string
	Parent         *Context
	ParentEntryPos *errors.Position
	SymbolTable    *SymbolTable
	InstanceID     int
}

func NewContext(displayName string, parent *Context, parentEntryPos *errors.Position) *Context {
	ctx := &Context{
		DisplayName:    displayName,
		Parent:         parent,
		ParentEntryPos: parentEntryPos,
	}

	if parent != nil {
		ctx.SymbolTable = NewSymbolTable(parent.SymbolTable)
		ctx.InstanceID = parent.InstanceID
	} else {
		ctx.SymbolTable = NewSymbolTable(nil)
	}

	return ctx
}

func GetInstanceID(ctx interface{}) int {
	if c, ok := ctx.(*Context); ok {
		return c.InstanceID
	}
	return 0
}
