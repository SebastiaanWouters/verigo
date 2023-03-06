package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/sebastiaanwouters/verigo/evaluator"
	"github.com/sebastiaanwouters/verigo/lexer"
	"github.com/sebastiaanwouters/verigo/object"
	"github.com/sebastiaanwouters/verigo/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	rMap := object.NewResultMap()
	c := make(chan int)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}
		evaluated := evaluator.Eval(program, env, rMap, c)
		if evaluated != nil {
		}
		value, ok := rMap.Get("a")
		if ok {
			fmt.Println("Saved Value: ", value.Inspect())
		}
	}
}

func Eval(input string, rMap *object.ResultMap, opChan chan int) {
	env := object.NewEnvironment()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	evaluator.Eval(program, env, rMap, opChan)
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
