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
	"chip-go/internal/interpreter"
	"chip-go/internal/lexer"
	"chip-go/internal/parser"
	"chip-go/internal/values"
	"reflect"
	"runtime"
)

type RoadRunner2 struct {
	interpreter   *interpreter.Interpreter
	globalContext *context.Context
}

func NewRoadRunner2() *RoadRunner2 {
	interp := interpreter.NewInterpreter()

	globalCtx := context.NewContext(constants.RR_CONTEXT_DISPLAY_NAME, nil, nil)

	roadRunner2 := &RoadRunner2{
		interpreter:   interp,
		globalContext: globalCtx,
	}

	values.SetGlobalInterpreter(interp)

	builtins.SetGlobalRoadRunner2(roadRunner2)

	builtinFuncs := builtins.GetBuiltins()
	for name, fn := range builtinFuncs {
		fn.SetContext(globalCtx)
		globalCtx.SymbolTable.Set(name, fn)
	}

	builtinConstants := builtins.GetConstants()
	for name, constant := range builtinConstants {
		constant.SetContext(globalCtx)
		globalCtx.SymbolTable.Set(name, constant)
	}

	return roadRunner2
}

func (rr *RoadRunner2) Run(filename, text string) (values.Value, error) {
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

	// Auto-garbage collect unassigned expression results to prevent memory leaks
	// Only GC if this is NOT a variable assignment (unassigned expressions)
	if node != nil {
		nodeTypeName := reflect.TypeOf(node).Elem().Name()

		if nodeTypeName != "VarAssignNode" {
			// Unassigned expression
			defer func() {
				runtime.GC()
				runtime.GC() // Double GC to be more aggressive
			}()
		}
	}

	return result.Value, nil
}

func (rr *RoadRunner2) GetGlobalContext() *context.Context {
	return rr.globalContext
}
