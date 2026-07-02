/*              .==,_
 *             .===,_`\
 *           .====,_ ` \      .====,__
 *     ---     .==-,`~. \           `:`.__,
 *      ---      `~~=-.  \           /^^^   ...always on the go!
 *        ---       `~~=. \         /
 *                     `~. \       /
 *                       ~. \____./
 *              jgs        `.=====)
 *                      ___.--~~~--.__
 *            ___\.--~~~              ~~~---.._|/
 *            ~~~"                             /
 *
 *
 * RR2 - internal/roadrunner/roadrunner.go
 *
 */

package roadrunner

import (
	"chip-go/internal/builtins"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"chip-go/internal/interpreter"
	"chip-go/internal/lexer"
	"chip-go/internal/orchestrator"
	"chip-go/internal/parser"
	"chip-go/internal/values"
	"fmt"
	"runtime/debug"
)

type RoadRunner2 struct {
	interpreter   *interpreter.Interpreter
	globalContext values.Ctx
}

var sharedInterpreter *interpreter.Interpreter

func NewRoadRunner2() *RoadRunner2 {
	if sharedInterpreter == nil {
		sharedInterpreter = interpreter.NewInterpreter()
	}

	values.SetGlobalInterpreter(sharedInterpreter)

	rr := newRR2(0, constants.RR_CONTEXT_DISPLAY_NAME)

	orch := orchestrator.New()
	orch.CreateMain(rr)

	orch.SetFactory(func(instanceID int) orchestrator.RR2Interface {
		return newRR2(instanceID, fmt.Sprintf("<Actor %d>", instanceID))
	})

	return rr
}

func recoverPanic(ctx values.Ctx, err *error) {
	if r := recover(); r != nil {
		if rtErr, ok := r.(*errors.RTError); ok {
			*err = rtErr
			return
		}

		var rrErr error

		if ctx.Trace != nil && ctx.Trace.Pos != nil {
			rrErr = errors.NewBaseError(ctx.Trace.Pos, nil, constants.PANIC_ERROR_TITLE, fmt.Sprintf("%v", r))
		} else {
			rrErr = fmt.Errorf("%s: %v", constants.PANIC_ERROR_TITLE, r)
		}

		*err = fmt.Errorf("%s\n\n%s", rrErr.Error(), debug.Stack())
	}
}

func (rr *RoadRunner2) Run(filename, text string) (retVal values.Value, retErr error) {
	defer recoverPanic(rr.globalContext, &retErr)

	lexer := lexer.NewLexer(filename, text)
	tokens, err := lexer.MakeTokens()

	if err != nil {
		return nil, err
	}

	parser := parser.NewParser(tokens)
	parseResult := parser.Parse()

	if parseResult.GetError() != nil {
		return nil, parseResult.GetError()
	}

	node := parseResult.GetNode()
	result := rr.interpreter.Visit(node, rr.globalContext)

	if result.Error != nil {
		return nil, result.Error
	}

	return result.Value, nil
}

func (rr *RoadRunner2) GetGlobalContext() values.Ctx {
	return rr.globalContext
}

func newRR2(instanceID int, displayName string) *RoadRunner2 {
	globalCtx := context.NewContext[values.Value](displayName, nil, nil)
	globalCtx.InstanceID = instanceID

	rr := &RoadRunner2{
		interpreter:   sharedInterpreter,
		globalContext: globalCtx,
	}

	for name, fn := range builtins.GetBuiltins() {
		fn.SetContext(globalCtx)
		globalCtx.SymbolTable.Set(name, fn)
	}

	for name, constant := range builtins.GetConstants() {
		constant.SetContext(globalCtx)
		globalCtx.SymbolTable.Set(name, constant)
	}

	globalCtx.SymbolTable.Set("CHIPRT", values.NewNumber(instanceID).SetContext(globalCtx))

	return rr
}
