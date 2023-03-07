package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/SebastiaanWouters/verigo/evaluator"
	"github.com/SebastiaanWouters/verigo/lexer"
	"github.com/SebastiaanWouters/verigo/object"
	"github.com/SebastiaanWouters/verigo/parser"
)

const PROMPT = ">> "

func monitorChan(c chan int) {
	for {
		<-c
	}
}

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	rMap := object.NewResultMap()
	c := make(chan int)
	go monitorChan(c)

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
