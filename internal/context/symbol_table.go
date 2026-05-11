/*
 *
 * RR2 - internal/context/symbol_table.go
 *
 */

package context

type SymbolTable[T any] struct {
	symbols map[string]T
	parent  *SymbolTable[T]
}

func NewSymbolTable[T any](parent *SymbolTable[T]) *SymbolTable[T] {
	return &SymbolTable[T]{
		symbols: make(map[string]T),
		parent:  parent,
	}
}

func (st *SymbolTable[T]) Get(name string) T {
	if value, exists := st.symbols[name]; exists {
		return value
	}

	if st.parent != nil {
		return st.parent.Get(name)
	}

	var zero T
	return zero
}

func (st *SymbolTable[T]) Exists(name string) bool {
	if _, exists := st.symbols[name]; exists {
		return true
	}

	if st.parent != nil {
		return st.parent.Exists(name)
	}

	return false
}

func (st *SymbolTable[T]) Set(name string, value T) {
	st.symbols[name] = value
}

// SetInScope sets a variable value in the scope where it already exists, or in
// the current scope if it doesn't exist anywhere in the scope chain.
func (st *SymbolTable[T]) SetInScope(name string, value T) {
	for current := st; current != nil; current = current.parent {
		if _, exists := current.symbols[name]; exists {
			current.symbols[name] = value
			return
		}
	}

	st.symbols[name] = value
}

func (st *SymbolTable[T]) Remove(name string) {
	delete(st.symbols, name)
}

// ForEach iterates over symbols in this scope only, not parents.
func (st *SymbolTable[T]) ForEach(fn func(name string, value T)) {
	for name, value := range st.symbols {
		fn(name, value)
	}
}

// SetParent rebinds the parent scope. Used during cross-actor value transfer so
// a sender-built closure snapshot can be attached to the receiver's globals after
// it crosses the actor boundary.
func (st *SymbolTable[T]) SetParent(parent *SymbolTable[T]) {
	st.parent = parent
}
