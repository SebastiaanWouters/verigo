package evaluator

import (
	"fmt"

	"github.com/SebastiaanWouters/verigo/ast"
	"github.com/SebastiaanWouters/verigo/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

var operationCount = 0

func Eval(node ast.Node, env *object.Environment, resultMap *object.ResultMap, opChan chan int) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node, env, resultMap, opChan)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env, resultMap, opChan)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env, resultMap, opChan)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env, resultMap, opChan)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env, resultMap, opChan)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right, opChan)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env, resultMap, opChan)
	case *ast.IfExpression:
		return evalIfExpression(node, env, resultMap, opChan)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env, resultMap, opChan)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env, resultMap, opChan)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.CallExpression:
		function := Eval(node.Function, env, resultMap, opChan)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env, resultMap, opChan)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args, env, resultMap, opChan)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment, rMap *object.ResultMap, opChan chan int) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env, rMap, opChan)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment, rMap *object.ResultMap, opChan chan int) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env, rMap, opChan)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalExpressions(
	exps []ast.Expression,
	env *object.Environment,
	rMap *object.ResultMap,
	opChan chan int,
) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env, rMap, opChan)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := utils[node.Value]; ok {
		return builtin
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment, rMap *object.ResultMap, opChan chan int) object.Object {
	condition := Eval(ie.Condition, env, rMap, opChan)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env, rMap, opChan)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env, rMap, opChan)
	} else {
		return NULL
	}
}
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalInfixExpression(operator string, left object.Object, right object.Object, c chan int) object.Object {
	switch {
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right, c)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(
	operator string,
	left, right object.Object,
) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
	c chan int,
) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "+":
		increaseOpCount(c)
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		increaseOpCount(c)
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		increaseOpCount(c)
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		increaseOpCount(c)
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		increaseOpCount(c)
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		increaseOpCount(c)
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		increaseOpCount(c)
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		increaseOpCount(c)
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value

	return &object.Integer{Value: -value}
}

func nativeBoolToBooleanObject(boolean bool) *object.Boolean {
	if boolean {
		return TRUE
	}
	return FALSE
}

func applyFunction(fn object.Object, args []object.Object, env *object.Environment, rMap *object.ResultMap, opChan chan int) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv, rMap, opChan)
		return unwrapReturnValue(evaluated)
	case *object.Save:
		return fn.Fn(args[0], args[1], env, rMap)
	case *object.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(
	fn *object.Function,
	args []object.Object,
) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func increaseOpCount(c chan int) {
	operationCount += 1
	if operationCount%10 == 0 {
		c <- 1
	}
}
