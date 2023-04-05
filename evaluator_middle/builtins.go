package evaluator_middle

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"

	"github.com/SebastiaanWouters/verigo/object"
)

func fib(n int64) int64 {
	var first int64 = 0
	var second int64 = 1
	if n == 0 {
		return first
	} else if n <= 2 {
		return second
	}
	for i := 2; int64(i) <= n; i++ {
		second = second + first
		first = second - first
	}
	return second
}

func isPrime(value int64) bool {
	for i := 2; i <= int(math.Floor(math.Sqrt(float64(value)))); i++ {
		if (value % int64(i)) == 0 {
			return false
		}
	}
	return value > 1
}

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Name: "len",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s",
					args[0].Type())
			}
		},
	},
	"pow": &object.Builtin{
		Name: "pow",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg1 := args[0].(type) {
			case *object.Integer:
				switch arg2 := args[1].(type) {
				case *object.Integer:
					return &object.Integer{Value: int64(math.Pow(float64(arg1.Value), float64(arg2.Value)))}
				default:
					return newError("argument to `pow` not supported, got %s",
						arg2.Type())
				}
			default:
				return newError("argument to `pow` not supported, got %s",
					arg1.Type())
			}

		},
	},
	"sqrt": &object.Builtin{
		Name: "sqrt",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				return &object.Integer{Value: int64(math.Sqrt(float64(arg.Value)))}
			default:
				return newError("argument to `pow` not supported, got %s",
					arg.Type())
			}

		},
	},
	"sin": &object.Builtin{
		Name: "sin",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				return &object.Integer{Value: int64(math.Sin(float64(arg.Value)))}
			default:
				return newError("argument to `pow` not supported, got %s",
					arg.Type())
			}

		},
	},
	"tan": &object.Builtin{
		Name: "tan",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				return &object.Integer{Value: int64(math.Tan(float64(arg.Value)))}
			default:
				return newError("argument to `pow` not supported, got %s",
					arg.Type())
			}

		},
	},
	"rand": &object.Builtin{
		Name: "rand",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				val, err := crand.Int(crand.Reader, big.NewInt(0xFFFF))
				if err != nil {
					return &object.Integer{Value: int64(0)}
				}
				return &object.Integer{Value: val.Int64()}
			default:
				return newError("argument to `pow` not supported, got %s",
					arg.Type())
			}

		},
	},
	"fib": &object.Builtin{
		Name: "fib",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				return &object.Integer{Value: fib(arg.Value)}
			default:
				return newError("argument to `pow` not supported, got %s",
					arg.Type())
			}

		},
	},
	"isPrime": &object.Builtin{
		Name: "isprime",
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				return &object.Boolean{Value: isPrime(arg.Value)}
			default:
				return newError("argument to `pow` not supported, got %s",
					arg.Type())
			}

		},
	},
	"print": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
}
