/*
 *
 * RR2 - internal/interpreter/interpreter.go
 *
 */

package interpreter

import (
	"chip-go/internal/ast"
	"chip-go/internal/constants"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"fmt"
	"reflect"
)

type Interpreter struct{}

func NewInterpreter() *Interpreter {
	return &Interpreter{}
}

func (i *Interpreter) getContext(ctx interface{}) *context.Context {
	if ctx == nil {
		return nil
	}
	if contextPtr, ok := ctx.(*context.Context); ok {
		return contextPtr
	}
	return nil
}

func (i *Interpreter) Visit(node ast.Node, ctx interface{}) *values.RuntimeResult {
	methodName := fmt.Sprintf("visit%s", reflect.TypeOf(node).Elem().Name())

	switch methodName {

	case "visitNumberNode":
		return i.visitNumberNode(node.(*ast.NumberNode), ctx)

	case "visitStringNode":
		return i.visitStringNode(node.(*ast.StringNode), ctx)

	case "visitListNode":
		return i.visitListNode(node.(*ast.ListNode), ctx)

	case "visitByteArrayNode":
		return i.visitByteArrayNode(node.(*ast.ByteArrayNode), ctx)

	case "visitVarAccessNode":
		return i.visitVarAccessNode(node.(*ast.VarAccessNode), ctx)

	case "visitVarAssignNode":
		return i.visitVarAssignNode(node.(*ast.VarAssignNode), ctx)

	case "visitVarUpdateNode":
		return i.visitVarUpdateNode(node.(*ast.VarUpdateNode), ctx)

	case "visitBinOpNode":
		return i.visitBinOpNode(node.(*ast.BinOpNode), ctx)

	case "visitUnaryOpNode":
		return i.visitUnaryOpNode(node.(*ast.UnaryOpNode), ctx)

	case "visitIfNode":
		return i.visitIfNode(node.(*ast.IfNode), ctx)

	case "visitForNode":
		return i.visitForNode(node.(*ast.ForNode), ctx)

	case "visitWhileNode":
		return i.visitWhileNode(node.(*ast.WhileNode), ctx)

	case "visitFuncDefNode":
		return i.visitFuncDefNode(node.(*ast.FuncDefNode), ctx)

	case "visitCallNode":
		return i.visitCallNode(node.(*ast.CallNode), ctx)

	case "visitReturnNode":
		return i.visitReturnNode(node.(*ast.ReturnNode), ctx)

	case "visitContinueNode":
		return i.visitContinueNode(node.(*ast.ContinueNode), ctx)

	case "visitBreakNode":
		return i.visitBreakNode(node.(*ast.BreakNode), ctx)

	default:
		panic(fmt.Sprintf("No visit method defined for %s", methodName))
	}
}

func (i *Interpreter) visitNumberNode(node *ast.NumberNode, ctx interface{}) *values.RuntimeResult {
	var value float64
	switch v := node.Token.Value.(type) {

	case int64:
		value = float64(v)

	case float64:
		value = v

	case int:
		value = float64(v)

	default:
		return values.NewRuntimeResult().Failure(errors.NewRTError(
			node.PosStart, node.PosEnd,
			fmt.Sprintf("Invalid number value: %v", node.Token.Value),
			ctx,
		))
	}

	return values.NewRuntimeResult().Success(
		values.NewNumber(value).SetContext(ctx).SetPos(node.PosStart, node.PosEnd),
	)
}

func (i *Interpreter) visitStringNode(node *ast.StringNode, ctx interface{}) *values.RuntimeResult {
	return values.NewRuntimeResult().Success(
		values.NewString(node.Token.Value.(string)).SetContext(ctx).SetPos(node.PosStart, node.PosEnd),
	)
}

func (i *Interpreter) visitListNode(node *ast.ListNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	elements := []values.Value{}

	for _, elementNode := range node.ElementNodes {
		elements = append(elements, res.Register(i.Visit(elementNode, ctx)))
		if res.ShouldReturn() {
			return res
		}
	}

	return res.Success(
		values.NewList(elements).SetContext(ctx).SetPos(node.PosStart, node.PosEnd),
	)
}

func (i *Interpreter) visitByteArrayNode(node *ast.ByteArrayNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	bytes := make([]byte, len(node.ElementNodes))

	for idx, elementNode := range node.ElementNodes {
		element := res.Register(i.Visit(elementNode, ctx))

		if res.ShouldReturn() {
			return res
		}

		// Validate that element is a number
		num, ok := element.(*values.Number)

		if !ok {
			return res.Failure(errors.NewRTError(
				node.PosStart, node.PosEnd,
				"Byte array elements must be numbers",
				ctx,
			))
		}

		// Validate byte range (0-255)
		byteVal := int(num.Value)

		if byteVal < 0 || byteVal > 255 {
			return res.Failure(errors.NewRTError(
				node.PosStart, node.PosEnd,
				"Byte values must be between 0 and 255",
				ctx,
			))
		}

		bytes[idx] = byte(byteVal)
	}

	return res.Success(
		values.NewBytes(bytes).SetContext(ctx).SetPos(node.PosStart, node.PosEnd),
	)
}

func (i *Interpreter) visitVarAccessNode(node *ast.VarAccessNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	varName := node.VarNameToken.Value.(string)
	context := i.getContext(ctx)

	if context == nil {
		return res.Failure(errors.NewRTError(
			node.PosStart, node.PosEnd,
			"Invalid context",
			ctx,
		))
	}

	value := context.SymbolTable.Get(varName)

	if value == nil {
		return res.Failure(errors.NewRTError(
			node.PosStart, node.PosEnd,
			fmt.Sprintf("'%s' is not defined", varName),
			ctx,
		))
	}

	if val, ok := value.(values.Value); ok {
		val = val.Copy().SetPos(node.PosStart, node.PosEnd).SetContext(ctx)

		return res.Success(val)
	}

	return res.Failure(errors.NewRTError(
		node.PosStart, node.PosEnd,
		fmt.Sprintf("Invalid value type for '%s'", varName),
		ctx,
	))
}

func (i *Interpreter) visitVarAssignNode(node *ast.VarAssignNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	varName := node.VarNameToken.Value.(string)
	value := res.Register(i.Visit(node.ValueNode, ctx))

	if res.ShouldReturn() {
		return res
	}

	context := i.getContext(ctx)

	if context == nil {
		return res.Failure(errors.NewRTError(
			node.PosStart, node.PosEnd,
			"Invalid context",
			ctx,
		))
	}

	context.SymbolTable.Set(varName, value)
	return res.Success(value)
}

func (i *Interpreter) visitVarUpdateNode(node *ast.VarUpdateNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	varName := node.VarNameToken.Value.(string)
	value := res.Register(i.Visit(node.ValueNode, ctx))

	if res.ShouldReturn() {
		return res
	}

	context := i.getContext(ctx)

	if context == nil {
		return res.Failure(errors.NewRTError(
			node.PosStart, node.PosEnd,
			"Invalid context",
			ctx,
		))
	}

	// Check if variable exists before updating
	if !context.SymbolTable.Exists(varName) {
		return res.Failure(errors.NewRTError(
			node.PosStart, node.PosEnd,
			fmt.Sprintf("'%s' is not defined", varName),
			ctx,
		))
	}

	context.SymbolTable.SetInScope(varName, value)

	return res.Success(value)
}

func (i *Interpreter) visitBinOpNode(node *ast.BinOpNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	left := res.Register(i.Visit(node.LeftNode, ctx))

	if res.ShouldReturn() {
		return res
	}

	right := res.Register(i.Visit(node.RightNode, ctx))

	if res.ShouldReturn() {
		return res
	}

	var result values.Value
	var err error

	switch node.OpToken.Type {

	case constants.TT_PLUS:
		result, err = left.AddedTo(right)

	case constants.TT_MINUS:
		result, err = left.SubbedBy(right)

	case constants.TT_MUL:
		result, err = left.MultedBy(right)

	case constants.TT_DIV:
		result, err = left.DivedBy(right)

	case constants.TT_POW:
		result, err = left.PowedBy(right)

	case constants.TT_EE:
		result, err = left.GetComparisonEe(right)

	case constants.TT_NE:
		result, err = left.GetComparisonNe(right)

	case constants.TT_LT:
		result, err = left.GetComparisonLt(right)

	case constants.TT_GT:
		result, err = left.GetComparisonGt(right)

	case constants.TT_LTE:
		result, err = left.GetComparisonLte(right)

	case constants.TT_GTE:
		result, err = left.GetComparisonGte(right)

	case constants.TT_KEYWORD:
		if node.OpToken.Value == "and" {
			result, err = left.AndedBy(right)
		} else if node.OpToken.Value == "or" {
			result, err = left.OredBy(right)
		}
	}

	if err != nil {
		return res.Failure(err)
	}

	return res.Success(result.SetPos(node.PosStart, node.PosEnd))
}

func (i *Interpreter) visitUnaryOpNode(node *ast.UnaryOpNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	number := res.Register(i.Visit(node.Node, ctx))

	if res.ShouldReturn() {
		return res
	}

	var result values.Value
	var err error

	switch node.OpToken.Type {

	case constants.TT_MINUS:
		result, err = number.MultedBy(values.NewNumber(-1))

	case constants.TT_PLUS:
		result = number

	case constants.TT_KEYWORD:
		if node.OpToken.Value == "not" {
			result, err = number.Notted()
		}
	}

	if err != nil {
		return res.Failure(err)
	}

	return res.Success(result.SetPos(node.PosStart, node.PosEnd))
}

func (i *Interpreter) visitIfNode(node *ast.IfNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	for _, ifCase := range node.Cases {
		conditionValue := res.Register(i.Visit(ifCase.Condition, ctx))

		if res.ShouldReturn() {
			return res
		}

		if conditionValue.IsTrue() {
			exprValue := res.Register(i.Visit(ifCase.Body, ctx))

			if res.ShouldReturn() {
				return res
			}

			if ifCase.ShouldReturnNull {
				return res.Success(values.NewNumber(constants.NUM_NUL))
			} else {
				return res.Success(exprValue)
			}
		}
	}

	if node.ElseCase != nil {
		elseValue := res.Register(i.Visit(node.ElseCase, ctx))

		if res.ShouldReturn() {
			return res
		}

		return res.Success(elseValue)
	}

	return res.Success(values.NewNumber(constants.NUM_NUL))
}

func (i *Interpreter) visitForNode(node *ast.ForNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	startValue := res.Register(i.Visit(node.StartValueNode, ctx))

	if res.ShouldReturn() {
		return res
	}

	endValue := res.Register(i.Visit(node.EndValueNode, ctx))

	if res.ShouldReturn() {
		return res
	}

	var stepValue values.Value = values.NewNumber(1)

	if node.StepValueNode != nil {
		stepValue = res.Register(i.Visit(node.StepValueNode, ctx))

		if res.ShouldReturn() {
			return res
		}
	}

	startNum, ok := startValue.(*values.Number)
	if !ok {
		return res.Failure(errors.NewRTError(
			node.StartValueNode.GetPosStart(), node.StartValueNode.GetPosEnd(),
			"Start value must be a number",
			ctx,
		))
	}

	endNum, ok := endValue.(*values.Number)
	if !ok {
		return res.Failure(errors.NewRTError(
			node.EndValueNode.GetPosStart(), node.EndValueNode.GetPosEnd(),
			"End value must be a number",
			ctx,
		))
	}

	stepNum, ok := stepValue.(*values.Number)
	if !ok {
		return res.Failure(errors.NewRTError(
			node.StepValueNode.GetPosStart(), node.StepValueNode.GetPosEnd(),
			"Step value must be a number",
			ctx,
		))
	}

	i_val := startNum.Value
	varName := node.VarNameToken.Value.(string)

	for {
		if stepNum.Value >= 0 && i_val >= endNum.Value {
			break
		}

		if stepNum.Value < 0 && i_val <= endNum.Value {
			break
		}

		context := i.getContext(ctx)

		if context != nil {
			context.SymbolTable.Set(varName, values.NewNumber(i_val))
		}

		res.Register(i.Visit(node.BodyNode, ctx))

		if res.ShouldReturn() && !res.LoopShouldContinue && !res.LoopShouldBreak {
			return res
		}

		if res.LoopShouldContinue {
			res.LoopShouldContinue = false
			i_val += stepNum.Value
			continue
		}

		if res.LoopShouldBreak {
			res.LoopShouldBreak = false
			break
		}

		i_val += stepNum.Value
	}

	context := i.getContext(ctx)

	if context != nil {
		context.SymbolTable.Remove(varName)
	}

	return res.Success(values.NewNumber(constants.NUM_NUL).SetContext(ctx).SetPos(node.PosStart, node.PosEnd))
}

func (i *Interpreter) visitWhileNode(node *ast.WhileNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	for {
		condition := res.Register(i.Visit(node.ConditionNode, ctx))

		if res.ShouldReturn() {
			return res
		}

		if !condition.IsTrue() {
			break
		}

		res.Register(i.Visit(node.BodyNode, ctx))

		if res.ShouldReturn() && !res.LoopShouldContinue && !res.LoopShouldBreak {
			return res
		}

		if res.LoopShouldContinue {
			res.LoopShouldContinue = false
			continue
		}

		if res.LoopShouldBreak {
			res.LoopShouldBreak = false
			break
		}

	}

	return res.Success(values.NewNumber(constants.NUM_NUL).SetContext(ctx).SetPos(node.PosStart, node.PosEnd))
}

func (i *Interpreter) visitFuncDefNode(node *ast.FuncDefNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	var funcName string
	if node.VarNameToken != nil {
		funcName = node.VarNameToken.Value.(string)
	}

	argNames := make([]string, len(node.ArgNameTokens))
	for i, token := range node.ArgNameTokens {
		argNames[i] = token.Value.(string)
	}

	funcValue := values.NewFunction(funcName, node.BodyNode, argNames, node.ShouldAutoReturn)
	funcValue.SetContext(ctx).SetPos(node.PosStart, node.PosEnd)

	if node.VarNameToken != nil {
		context := i.getContext(ctx)

		if context != nil {
			context.SymbolTable.Set(funcName, funcValue)
		}
	}

	return res.Success(funcValue)
}

func (i *Interpreter) visitCallNode(node *ast.CallNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()
	args := []values.Value{}

	valueToCall := res.Register(i.Visit(node.NodeToCall, ctx))
	if res.ShouldReturn() {
		return res
	}
	valueToCall = valueToCall.Copy()

	for _, argNode := range node.ArgNodes {
		args = append(args, res.Register(i.Visit(argNode, ctx)))

		if res.ShouldReturn() {
			return res
		}
	}

	returnValue := res.Register(valueToCall.Execute(args))

	if res.ShouldReturn() {
		return res
	}

	returnValue = returnValue.Copy().SetPos(node.PosStart, node.PosEnd).SetContext(ctx)

	return res.Success(returnValue)
}

func (i *Interpreter) visitReturnNode(node *ast.ReturnNode, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	var value values.Value = values.NewNumber(constants.NUM_NUL)

	if node.NodeToReturn != nil {
		value = res.Register(i.Visit(node.NodeToReturn, ctx))

		if res.ShouldReturn() {
			return res
		}
	}

	return res.SuccessReturn(value)
}

func (i *Interpreter) visitContinueNode(node *ast.ContinueNode, ctx interface{}) *values.RuntimeResult {
	return values.NewRuntimeResult().SuccessContinue()
}

func (i *Interpreter) visitBreakNode(node *ast.BreakNode, ctx interface{}) *values.RuntimeResult {
	return values.NewRuntimeResult().SuccessBreak()
}
