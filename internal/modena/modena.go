/*
 *
 * Chippy - internal/modena/modena.go
 *
 */

package modena

import (
	"chip-go/internal/builtins"
	"chip-go/internal/bytecode/serializer"
	"chip-go/internal/compiler"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"chip-go/internal/vm"
	"fmt"
	"runtime/debug"
)

type Modena struct {
	globalContext values.Ctx
	schema        *context.Schema
}

func New() *Modena {
	mod := newInstance(0, constants.CONTEXT_DISPLAY_NAME)

	orch := orchestrator.New()
	orch.CreateMain(mod)

	orch.SetFactory(func(instanceID int) orchestrator.Modena {
		return newInstance(instanceID, fmt.Sprintf("<Actor %d>", instanceID))
	})

	return mod
}

func newInstance(instanceID int, displayName string) *Modena {
	schema := context.NewSchema()

	globalCtx := context.NewContext[values.Value](displayName, nil, nil)
	globalCtx.InstanceID = instanceID
	globalCtx.Globals = context.NewGlobals[values.Value](schema)

	for name, fn := range builtins.GetBuiltins() {
		globalCtx.Globals.SetByName(name, fn)
	}

	for name, constant := range builtins.GetConstants() {
		globalCtx.Globals.SetByName(name, constant)
	}

	// Set CHIPRT
	globalCtx.Globals.SetByName("CHIPRT", values.NewNumber(instanceID))

	return &Modena{globalContext: globalCtx, schema: schema}
}

func (e *Modena) Run(filename, text string) (retVal values.Value, retErr error) {
	defer recoverPanic(e.globalContext, &retErr, serializer.LooksLikeBytecode([]byte(text)))

	chunk, err := compiler.DecodeOrCompile(filename, []byte(text))

	if err != nil {
		return values.Value{}, err
	}

	chunk.LinkGlobals(e.schema)

	return vm.NewVM(chunk, e.globalContext).Run()
}

func (e *Modena) GetGlobalContext() values.Ctx {
	return e.globalContext
}

// recoverPanic turns a Go fault into an RTError rather than crashing the host.
// A raised RTError passes through unchanged.
func recoverPanic(ctx values.Ctx, err *error, fromBytecode bool) {
	recovered := recover()

	if recovered == nil {
		return
	}

	if rtErr, ok := recovered.(*errors.RTError); ok {
		*err = rtErr

		return
	}

	var pos *errors.Position

	if ctx.Trace != nil {
		pos = ctx.Trace.Pos
	}

	if fromBytecode {
		*err = errors.NewBaseError(pos, nil, constants.DAMAGED_ERROR_TITLE, constants.E_DAMAGED_PROGRAM)

		return
	}

	base := errors.NewBaseError(pos, nil, constants.PANIC_ERROR_TITLE, fmt.Sprintf("%v", recovered))
	*err = fmt.Errorf("%s\n\n%s", base.Error(), debug.Stack())
}
