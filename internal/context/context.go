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
	IsTemporary    bool // Track if this context should be cleaned up after use
}

func NewContext(displayName string, parent *Context, parentEntryPos *errors.Position) *Context {
	ctx := &Context{
		DisplayName:    displayName,
		Parent:         parent,
		ParentEntryPos: parentEntryPos,
	}

	if parent != nil {
		ctx.SymbolTable = NewSymbolTable(parent.SymbolTable)
	} else {
		ctx.SymbolTable = NewSymbolTable(nil)
	}

	return ctx
}

// Cleanup explicitly clears the context and breaks references to help GC
func (c *Context) Cleanup() {
	if c.SymbolTable != nil {
		c.SymbolTable.Clear()
		c.SymbolTable = nil
	}
	// Break parent reference to prevent memory leaks
	c.Parent = nil
	c.ParentEntryPos = nil
}
