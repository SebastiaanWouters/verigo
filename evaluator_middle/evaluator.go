package evaluator_middle

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

func Eval(node ast.Node, env *object.Environment, opCount *int) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalProgram(node, env, opCount)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env, opCount)
	// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env, opCount)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env, opCount)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env, opCount)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right, opCount)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env, opCount)
	case *ast.IfExpression:
		return evalIfExpression(node, env, opCount)
	case *ast.ForExpression:
		return evalForExpression(node, env, opCount)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env, opCount)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env, opCount)
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
		function := Eval(node.Function, env, opCount)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env, opCount)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args, env, opCount)
	}

	return nil
}

func evalProgram(program *ast.Program, env *object.Environment, opCount *int) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env, opCount)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment, opCount *int) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env, opCount)
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
	opCount *int,
) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env, opCount)
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
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	return newError("identifier not found: " + node.Value)
}

func evalForExpression(ie *ast.ForExpression, env *object.Environment, opCount *int) object.Object {
	Eval(&ie.Variable, env, opCount)
	condition := Eval(ie.Condition, env, opCount)
	if isError(condition) {
		return condition
	}
	for isTruthy(condition) {
		Eval(ie.Loop, env, opCount)
		Eval(&ie.Update, env, opCount)
		condition = Eval(ie.Condition, env, opCount)
		if isError(condition) {
			return condition
		}
	}
	return NULL
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment, opCount *int) object.Object {
	condition := Eval(ie.Condition, env, opCount)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env, opCount)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env, opCount)
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

func evalInfixExpression(operator string, left object.Object, right object.Object, c *int) object.Object {
	switch {
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right, c)
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
	c *int,
) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	*c += 1
	return &object.String{Value: leftVal + rightVal}
}

func evalIntegerInfixExpression(
	operator string,
	left, right object.Object,
	c *int,
) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "+":
		*c += 1
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		*c += 1
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		*c += 1
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		*c += 1
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		*c += 1
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		*c += 1
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		*c += 1
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		*c += 1
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

func applyFunction(fn object.Object, args []object.Object, env *object.Environment, c *int) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv, c)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		switch fn.Name {
		case "isprime":
			*c += 1
		case "sin":
			*c += 1
		case "tan":
			*c += 1
		case "rand":
			*c += 1
		case "pow":
			*c += 1
		case "sqrt":
			*c += 1
		case "len":
			*c += 1
		case "fib":
			*c += 1
		}
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
