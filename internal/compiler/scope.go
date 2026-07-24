/*
 *
 * Modena - internal/compiler/scope.go
 *
 */

package compiler

import "chip-go/internal/bytecode"

type localVar struct {
	name string
	slot int
}

type funcScope struct {
	active   []localVar
	slots    int
	names    []string
	upvalues []bytecode.UpvalueDesc
}

func (c *Compiler) currentScope() *funcScope {
	return c.scopes[len(c.scopes)-1]
}

func (c *Compiler) addLocal(name string) int {
	scope := c.currentScope()
	slot := scope.slots
	scope.slots++
	scope.names = append(scope.names, name)
	scope.active = append(scope.active, localVar{name: name, slot: slot})

	return slot
}

func (c *Compiler) resolveLocal(name string) int {
	active := c.currentScope().active

	for i := len(active) - 1; i >= 0; i-- {
		if active[i].name == name {
			return active[i].slot
		}
	}

	return -1
}

func (c *Compiler) beginScope() int {
	c.blockDepth++

	return len(c.currentScope().active)
}

func (c *Compiler) endScope(mark int) {
	c.blockDepth--

	scope := c.currentScope()
	scope.active = scope.active[:mark]
}

func (c *Compiler) resolveUpvalue(scopeIdx int, name string) int {
	if scopeIdx == 0 {
		return -1
	}

	enclosing := c.scopes[scopeIdx-1]

	for i := len(enclosing.active) - 1; i >= 0; i-- {
		if enclosing.active[i].name == name {
			return c.addUpvalue(scopeIdx, bytecode.UpvalueDesc{FromLocal: true, Index: enclosing.active[i].slot, Name: name})
		}
	}

	if up := c.resolveUpvalue(scopeIdx-1, name); up >= 0 {
		return c.addUpvalue(scopeIdx, bytecode.UpvalueDesc{FromLocal: false, Index: up, Name: name})
	}

	return -1
}

func (c *Compiler) addUpvalue(scopeIdx int, desc bytecode.UpvalueDesc) int {
	scope := c.scopes[scopeIdx]

	for i, existing := range scope.upvalues {
		if existing == desc {
			return i
		}
	}

	scope.upvalues = append(scope.upvalues, desc)

	return len(scope.upvalues) - 1
}
