/*
 *
 * RR2 - internal/context/symbol_table.go
 *
 */

package context

type SymbolTable struct {
	symbols map[string]interface{}
	parent  *SymbolTable
}

func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		symbols: make(map[string]interface{}),
		parent:  parent,
	}
}

func (st *SymbolTable) Get(name string) interface{} {
	value, exists := st.symbols[name]
	if !exists && st.parent != nil {
		return st.parent.Get(name)
	}
	return value
}

func (st *SymbolTable) Exists(name string) bool {
	_, exists := st.symbols[name]
	if !exists && st.parent != nil {
		return st.parent.Exists(name)
	}
	return exists
}

func (st *SymbolTable) Set(name string, value interface{}) {
	st.symbols[name] = value
}

// SetInScope sets a variable value in the scope where it already exists,
// or in the current scope if it doesn't exist anywhere in the scope chain
func (st *SymbolTable) SetInScope(name string, value interface{}) {
	current := st
	for current != nil {
		if _, exists := current.symbols[name]; exists {
			current.symbols[name] = value
			return
		}
		current = current.parent
	}
	// Variable doesn't exist in any scope, set it in current scope
	st.symbols[name] = value
}

func (st *SymbolTable) Remove(name string) {
	delete(st.symbols, name)
}

// Clear removes all symbols and breaks parent reference for cleanup
func (st *SymbolTable) Clear() {
	for k := range st.symbols {
		delete(st.symbols, k)
	}
	st.parent = nil
}
