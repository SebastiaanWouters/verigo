package evaluator

import (
	"fmt"
	"math"
	"math/big"

	"github.com/SebastiaanWouters/verigo/object"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
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
	"isPrime": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				return &object.Boolean{Value: big.NewInt(arg.Value).ProbablyPrime(0)}
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

var utils = map[string]*object.Save{
	"save": &object.Save{
		Fn: func(name object.Object, id object.Object, env *object.Environment, rMap *object.ResultMap) object.Object {
			if name.Type() == object.STRING_OBJ {
				rMap.Set(name.Inspect(), id)
				return NULL
			} else {
				return newError("arguments to `save` not supported, got %s",
					name.Type())
			}
		},
	},
}
