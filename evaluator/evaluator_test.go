package evaluator_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/SebastiaanWouters/verigo/evaluator"
	"github.com/SebastiaanWouters/verigo/lexer"
	"github.com/SebastiaanWouters/verigo/object"
	"github.com/SebastiaanWouters/verigo/parser"
	"github.com/SebastiaanWouters/verigo/repl"
)

const (
	AMOUNT       = 10000
	ADDITION     = 0
	SUBSTRACTION = 1
	MULTIPLY     = 2
	DIVIDE       = 3
	ST           = 4
	GT           = 5
	EQ           = 6
	NEQ          = 7
	PRIME        = 8
	SIN          = 9
	TAN          = 10
	RAND         = 11
	POW          = 12
	SQRT         = 13
	LEN          = 14
	FIB          = 15
	CONCAT       = 16
)

var WEIGHTS = map[int]float32{
	ADDITION:     1.0047,
	SUBSTRACTION: 1.0,
	MULTIPLY:     1.0025,
	DIVIDE:       1.0147,
	ST:           1,
	GT:           1,
	EQ:           1,
	NEQ:          1,
	PRIME:        1.5261,
	SIN:          1.4841,
	TAN:          1.4907,
	RAND:         3.5527,
	POW:          1.8782,
	SQRT:         1.43,
	LEN:          1.5093,
	FIB:          2.4969,
	CONCAT:       1, //TODO
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{"if (10 > 1) { return 10; }", 10},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return 10;
  }

  return 1;
}
`,
			10,
		},
		{
			`
let f = fn(x) {
  return x;
  x + 10;
};
f(10);`,
			10,
		},
		{
			`
let f = fn(x) {
   let result = x + 10;
   return result;
   return 10;
};
f(10);`,
			20,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"true + false + true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return true + false;
  }

  return 1;
}
`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`{"name": "Monkey"}[fn(x) { x }];`,
			"unusable as hash key: FUNCTION",
		},
		{
			`999[1]`,
			"index operator not supported: INTEGER",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestEnclosingEnvironments(t *testing.T) {
	input := `
let first = 10;
let second = 10;
let third = 10;

let ourFunction = fn(first) {
  let second = 20;

  first + second + third;
};

ourFunction(20) + first + second;`

	testIntegerObject(t, testEval(input), 70)
}

func TestClosures(t *testing.T) {
	input := `
let newAdder = fn(x) {
  fn(y) { x + y };
};

let addTwo = newAdder(2);
addTwo(2);`

	testIntegerObject(t, testEval(input), 4)
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestSimpleAddition(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generateInfix(AMOUNT, "+", object.INTEGER_OBJ)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "additions resulted in a count of: ", counter)
}

func TestAdvancedAddition(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generateInfix(AMOUNT, "+", object.INTEGER_OBJ)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "additions resulted in a count of: ", counter)
}

func TestSimpleSubstraction(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generateInfix(AMOUNT, "-", object.INTEGER_OBJ)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "substractions resulted in a count of: ", counter)
}

func TestAdvancedSubstraction(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generateInfix(AMOUNT, "-", object.INTEGER_OBJ)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "substractions resulted in a count of: ", counter)
}

func TestSimpleMultiply(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generateInfix(AMOUNT, "*", object.INTEGER_OBJ)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "multiplications resulted in a count of: ", counter)
}

func TestAdvancedMultiply(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generateInfix(AMOUNT, "*", object.INTEGER_OBJ)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "multiplications resulted in a count of: ", counter)
}

func TestSimpleDivide(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generateInfix(AMOUNT, "/", object.INTEGER_OBJ)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "divisions resulted in a count of: ", counter)
}

func TestAdvancedDivide(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generateInfix(AMOUNT, "/", object.INTEGER_OBJ)

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "divisions resulted in a count of: ", counter)
}

func TestSimplePrime(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generatePrefix(AMOUNT, 1, "isPrime", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "primes resulted in a count of: ", counter)
}

func TestAdvancedPrime(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generatePrefix(AMOUNT, 1, "isPrime", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "primes resulted in a count of: ", counter)
}

func TestSimpleSin(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generatePrefix(AMOUNT, 1, "sin", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "sines resulted in a count of: ", counter)
}

func TestAdvancedSin(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generatePrefix(AMOUNT, 1, "sin", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "sines resulted in a count of: ", counter)
}

func TestSimpleTan(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generatePrefix(AMOUNT, 1, "tan", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "tans resulted in a count of: ", counter)
}

func TestAdvancedTan(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generatePrefix(AMOUNT, 1, "tan", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "tans resulted in a count of: ", counter)
}

func TestSimpleRand(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generatePrefix(AMOUNT, 1, "rand", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "rands resulted in a count of: ", counter)
}

func TestAdvancedRand(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generatePrefix(AMOUNT, 1, "rand", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	start := time.Now()
	repl.EvalParsed(program, env, rChan, opChan)
	elapsed := time.Since(start)
	work := counter / float32(elapsed.Microseconds())

	t.Log("running", AMOUNT, "rands resulted in a count of: ", counter)
	t.Log("Performed", work, "work units per time unit")
}

func TestSimplePow(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generatePrefix(AMOUNT, 2, "pow", []object.ObjectType{object.INTEGER_OBJ, object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "pows resulted in a count of: ", counter)
}

func TestAdvancedPow(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generatePrefix(AMOUNT, 2, "pow", []object.ObjectType{object.INTEGER_OBJ, object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	start := time.Now()
	repl.EvalParsed(program, env, rChan, opChan)
	elapsed := time.Since(start)
	work := counter / float32(elapsed.Microseconds())

	t.Log("running", AMOUNT, "pows resulted in a count of: ", counter)
	t.Log("Performed", work, "work units per time unit")
}

func TestSimpleSqrt(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generatePrefix(AMOUNT, 1, "sqrt", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "sqrts resulted in a count of: ", counter)
}

func TestAdvancedSqrt(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generatePrefix(AMOUNT, 1, "sqrt", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	start := time.Now()
	repl.EvalParsed(program, env, rChan, opChan)
	elapsed := time.Since(start)
	work := counter / float32(elapsed.Microseconds())

	t.Log("running", AMOUNT, "sqrt resulted in a count of: ", counter)
	t.Log("Performed", work, "work units per time unit")
}

func TestSimpleLen(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generatePrefix(AMOUNT, 1, "len", []object.ObjectType{object.STRING_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "lens resulted in a count of: ", counter)
}

func TestAdvancedLen(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generatePrefix(AMOUNT, 1, "len", []object.ObjectType{object.STRING_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	start := time.Now()
	repl.EvalParsed(program, env, rChan, opChan)
	elapsed := time.Since(start)
	work := counter / float32(elapsed.Microseconds())

	t.Log("running", AMOUNT, "lens resulted in a count of: ", counter)
	t.Log("Performed", work, "work units per time unit")
}

func TestSimpleFib(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			<-opChan
			counter = counter + 1
		}
	}()

	input := generatePrefix(AMOUNT, 1, "fib", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "fibs resulted in a count of: ", counter)
}

func TestAdvancedFib(t *testing.T) {
	var counter float32 = 0
	opChan := make(chan int)
	rChan := make(chan object.Result)

	go func() {
		for {
			op := <-opChan
			counter += WEIGHTS[op]
		}
	}()

	input := generatePrefix(AMOUNT, 1, "fib", []object.ObjectType{object.INTEGER_OBJ})

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	repl.EvalParsed(program, env, rChan, opChan)

	t.Log("running", AMOUNT, "fibs resulted in a count of: ", counter)
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()
	opChan := make(chan int)
	rChan := make(chan object.Result)
	go opChanMonitor(opChan)
	go rChanMonitor(rChan)
	return evaluator.Eval(program, env, rChan, opChan)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj.Type() != object.NULL_OBJ {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func opChanMonitor(c chan int) {
	for {
		<-c
	}
}
func rChanMonitor(c chan object.Result) {
	for {
		<-c
	}
}

// Benchmarks
func BenchmarkFile(b *testing.B) {
	opChan := make(chan int)
	rChan := make(chan object.Result)
	go opChanMonitor(opChan)
	go rChanMonitor(rChan)

	addStr := generateInfix(AMOUNT, "+", object.INTEGER_OBJ)
	subStr := generateInfix(AMOUNT, "-", object.INTEGER_OBJ)
	mulStr := generateInfix(AMOUNT, "*", object.INTEGER_OBJ)
	divStr := generateInfix(AMOUNT, "/", object.INTEGER_OBJ)
	primeStr := generatePrefix(AMOUNT, 1, "isPrime", []object.ObjectType{object.INTEGER_OBJ})
	sinStr := generatePrefix(AMOUNT, 1, "sin", []object.ObjectType{object.INTEGER_OBJ})
	tanStr := generatePrefix(AMOUNT, 1, "tan", []object.ObjectType{object.INTEGER_OBJ})
	randStr := generatePrefix(AMOUNT, 1, "rand", []object.ObjectType{object.INTEGER_OBJ})
	fibStr := generatePrefix(AMOUNT, 1, "fib", []object.ObjectType{object.INTEGER_OBJ})
	powStr := generatePrefix(AMOUNT, 2, "pow", []object.ObjectType{object.INTEGER_OBJ, object.INTEGER_OBJ})
	lenStr := generatePrefix(AMOUNT, 1, "len", []object.ObjectType{object.STRING_OBJ})
	sqrtStr := generatePrefix(AMOUNT, 1, "sqrt", []object.ObjectType{object.INTEGER_OBJ})
	concatStr := generateInfix(AMOUNT, "+", object.STRING_OBJ)

	b.Run("addition", func(b *testing.B) {
		l := lexer.New(addStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()

		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[ADDITION], "work units or", float32(b.N*AMOUNT)*WEIGHTS[ADDITION]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("substraction", func(b *testing.B) {
		l := lexer.New(subStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[SUBSTRACTION], "work units or", float32(b.N*AMOUNT)*WEIGHTS[SUBSTRACTION]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("multiply", func(b *testing.B) {
		l := lexer.New(mulStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[MULTIPLY], "work units or", float32(b.N*AMOUNT)*WEIGHTS[MULTIPLY]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("divide", func(b *testing.B) {
		l := lexer.New(divStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[DIVIDE], "work units or", float32(b.N*AMOUNT)*WEIGHTS[DIVIDE]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("isPrime", func(b *testing.B) {
		l := lexer.New(primeStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[PRIME], "work units or", float32(b.N*AMOUNT)*WEIGHTS[PRIME]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("sin", func(b *testing.B) {
		l := lexer.New(sinStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[SIN], "work units or", float32(b.N*AMOUNT)*WEIGHTS[SIN]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("tan", func(b *testing.B) {
		l := lexer.New(tanStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[TAN], "work units or", float32(b.N*AMOUNT)*WEIGHTS[TAN]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("rand", func(b *testing.B) {
		l := lexer.New(randStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[RAND], "work units or", float32(b.N*AMOUNT)*WEIGHTS[RAND]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("fib", func(b *testing.B) {
		l := lexer.New(fibStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[FIB], "work units or", float32(b.N*AMOUNT)*WEIGHTS[FIB]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("pow", func(b *testing.B) {
		l := lexer.New(powStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[POW], "work units or", float32(b.N*AMOUNT)*WEIGHTS[POW]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("len", func(b *testing.B) {
		l := lexer.New(lenStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[LEN], "work units or", float32(b.N*AMOUNT)*WEIGHTS[LEN]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("sqrt", func(b *testing.B) {
		l := lexer.New(sqrtStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()
		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[SQRT], "work units or", float32(b.N*AMOUNT)*WEIGHTS[SQRT]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})
	b.Run("concat", func(b *testing.B) {
		l := lexer.New(concatStr)
		p := parser.New(l)
		program := p.ParseProgram()
		env := object.NewEnvironment()

		for i := 0; i < b.N; i++ {
			repl.EvalParsed(program, env, rChan, opChan)
		}
		fmt.Println("SIMPLE: Received", float32(b.N*AMOUNT), "work units, or", float32(b.N*AMOUNT)/float32(b.Elapsed().Microseconds()), "work units per microsecond")
		fmt.Println("ADVANCED: Received", float32(b.N)*WEIGHTS[ADDITION], "work units or", float32(b.N*AMOUNT)*WEIGHTS[ADDITION]/float32(b.Elapsed().Microseconds()), "work units per microsecond")
	})

}

func generateRandomNumber(from int, to int) string {
	return strconv.Itoa(rand.Intn(to) + from)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRandomString(size int) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func addLine(input string, line string) string {
	if input == "" {
		return line
	}
	return input + " \n" + line
}

func generateInfix(amount int, operator string, argtype object.ObjectType) string {
	program := ""
	if argtype == object.INTEGER_OBJ {
		for i := 0; i < amount; i++ {
			num1 := generateRandomNumber(100, 2000)
			num2 := generateRandomNumber(100, 2000)
			newLine := num1 + operator + num2
			program = addLine(program, newLine)
		}
	} else if argtype == object.STRING_OBJ {
		for i := 0; i < amount; i++ {
			str1 := generateRandomString(20)
			str2 := generateRandomString(20)
			newLine := "\"" + str1 + "\"" + operator + "\"" + str2 + "\""
			program = addLine(program, newLine)
		}
	}

	return program
}

func generatePrefix(amount int, args int, operator string, argtypes []object.ObjectType) string {
	program := ""
	for i := 0; i < amount; i++ {
		newLine := generatePrefixOp(operator, args, argtypes)
		program = addLine(program, newLine)
	}
	return program
}

func generatePrefixOp(operator string, args int, argtypes []object.ObjectType) string {
	op := operator + "("
	for i := 0; i < args; i++ {
		if argtypes[i] == object.INTEGER_OBJ {
			if i > 0 {
				//not first argument
				if operator == "pow" {
					op = op + "," + generateRandomNumber(1, 7)
				} else {
					op = op + "," + generateRandomNumber(100, 1000)
				}
			} else {
				//first argument
				if operator == "pow" {
					op = op + generateRandomNumber(1, 20)
				} else {
					op = op + generateRandomNumber(100, 1000)
				}
			}
		} else if argtypes[i] == object.STRING_OBJ {
			if i > 0 {
				op = op + "," + "\"" + generateRandomString(rand.Intn(20)) + "\""
			} else {
				op = op + "\"" + generateRandomString(rand.Intn(20)) + "\""
			}
		}
	}
	op += ")"
	return op
}
