/*
 *
 * Modena - internal/modena/modena.go
 *
 */

package modena

import (
	"chip-go/internal/builtins"
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
	globalCtx := context.NewContext[values.Value](displayName, nil, nil)
	globalCtx.InstanceID = instanceID

	for name, fn := range builtins.GetBuiltins() {
		fn.SetContext(globalCtx)
		globalCtx.SymbolTable.Set(name, fn)
	}

	for name, constant := range builtins.GetConstants() {
		constant.SetContext(globalCtx)
		globalCtx.SymbolTable.Set(name, constant)
	}

	// Set CHIPRT
	globalCtx.SymbolTable.Set("CHIPRT", values.NewNumber(instanceID).SetContext(globalCtx))

	return &Modena{globalContext: globalCtx}
}

func (e *Modena) Run(filename, text string) (retVal values.Value, retErr error) {
	defer recoverPanic(e.globalContext, &retErr)

	chunk, err := compiler.DecodeOrCompile(filename, []byte(text))

	if err != nil {
		return nil, err
	}

	return vm.NewVM(chunk, e.globalContext).Run()
}

func (e *Modena) GetGlobalContext() values.Ctx {
	return e.globalContext
}

// recoverPanic turns a Go fault into an RTError rather than crashing the host.
// A raised RTError passes through unchanged.
func recoverPanic(ctx values.Ctx, err *error) {
	recovered := recover()

	if recovered == nil {
		return
	}

	if rtErr, ok := recovered.(*errors.RTError); ok {
		*err = rtErr

		return
	}

	if ctx.Trace != nil && ctx.Trace.Pos != nil {
		base := errors.NewBaseError(ctx.Trace.Pos, nil, constants.PANIC_ERROR_TITLE, fmt.Sprintf("%v", recovered))
		*err = fmt.Errorf("%s\n\n%s", base.Error(), debug.Stack())

		return
	}

	*err = fmt.Errorf("%s: %v\n\n%s", constants.PANIC_ERROR_TITLE, recovered, debug.Stack())
}
