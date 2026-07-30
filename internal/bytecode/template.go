/*
 *
 * Chippy - internal/bytecode/template.go
 *
 */

package bytecode

type UpvalueDesc struct {
	FromLocal bool
	Index     int
	Name      string
}

type FunctionTemplate struct {
	Name     string
	Chunk    *Chunk
	Arity    int
	Upvalues []UpvalueDesc
}
