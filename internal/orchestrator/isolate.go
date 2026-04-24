/*
 *
 * RR2 - internal/orchestrator/isolate.go
 *
 */

package orchestrator

import (
	"chip-go/internal/context"
	"chip-go/internal/values"
)

// IsolateForTransfer snapshots every closure's captured environment onto a fresh
// detached context so calls on the receiving actor don't race the sender's writes.
// cycles dedupes shared captures and terminates recursion.
func IsolateForTransfer(v values.Value, cycles map[*context.Context]*context.Context) {
	if v == nil {
		return
	}

	switch val := v.(type) {
	case *values.Function:
		isolateTransferFunction(val, cycles)
	case *values.List:
		for _, e := range val.Elements {
			IsolateForTransfer(e, cycles)
		}
	case *values.Map:
		for _, e := range val.Entries {
			IsolateForTransfer(e, cycles)
		}
	}
}

func isolateTransferFunction(fn *values.Function, cycles map[*context.Context]*context.Context) {
	origCtx, ok := fn.GetContext().(*context.Context)

	if !ok || origCtx == nil {
		return
	}

	if existing, found := cycles[origCtx]; found {
		fn.SetContext(existing)
		return
	}

	detached := context.NewContext("<sent>", nil, nil)
	cycles[origCtx] = detached

	var chain []*context.Context

	for c := origCtx; c != nil; c = c.Parent {
		chain = append(chain, c)
	}

	// Outer-to-inner so inner names shadow outer.
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].SymbolTable == nil {
			continue
		}

		chain[i].SymbolTable.ForEach(func(name string, raw interface{}) {
			if name == "CHIPRT" {
				return
			}

			val, ok := raw.(values.Value)

			if !ok {
				return
			}

			// Builtins are stateless; the receiver reaches its own
			// copies through its globals once the detached context
			// is parented.
			if _, isBuiltin := val.(*values.BuiltInFunction); isBuiltin {
				return
			}

			snap := val.Copy()
			IsolateForTransfer(snap, cycles)
			detached.SymbolTable.Set(name, snap)
		})
	}

	fn.SetContext(detached)
}

// DeepRebindContext stamps ctx onto every non-function value in the graph.
// Functions are skipped; their context is the closure capture, handled by
// IsolateForTransfer.
func DeepRebindContext(v values.Value, ctx interface{}) {
	if v == nil {
		return
	}

	switch val := v.(type) {
	case *values.Function, *values.BuiltInFunction:

	case *values.List:
		v.SetContext(ctx)

		for _, e := range val.Elements {
			DeepRebindContext(e, ctx)
		}

	case *values.Map:
		v.SetContext(ctx)

		for _, e := range val.Entries {
			DeepRebindContext(e, ctx)
		}

	default:
		v.SetContext(ctx)
	}
}

// BindValuesToGlobals parents every detached closure context onto the receiver's
// globals so sent functions can resolve builtins and see the receiver's InstanceID.
func BindValuesToGlobals(items []values.Value, globals *context.Context) {
	if globals == nil {
		return
	}

	visited := make(map[*context.Context]bool)

	for _, item := range items {
		bindValueToGlobals(item, globals, visited)
	}
}

func bindValueToGlobals(v values.Value, globals *context.Context, visited map[*context.Context]bool) {
	if v == nil {
		return
	}

	switch val := v.(type) {
	case *values.Function:
		c, ok := val.GetContext().(*context.Context)

		if !ok || c == nil || visited[c] {
			return
		}

		visited[c] = true
		c.Parent = globals
		c.InstanceID = globals.InstanceID

		if c.SymbolTable != nil && globals.SymbolTable != nil {
			c.SymbolTable.SetParent(globals.SymbolTable)
		}

		if c.SymbolTable != nil {
			c.SymbolTable.ForEach(func(name string, raw interface{}) {
				if inner, ok := raw.(values.Value); ok {
					bindValueToGlobals(inner, globals, visited)
				}
			})
		}

	case *values.List:
		for _, e := range val.Elements {
			bindValueToGlobals(e, globals, visited)
		}

	case *values.Map:
		for _, e := range val.Entries {
			bindValueToGlobals(e, globals, visited)
		}
	}
}
