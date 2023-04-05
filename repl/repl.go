package repl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/SebastiaanWouters/verigo/evaluator_middle"
	"github.com/SebastiaanWouters/verigo/parser"

	"github.com/SebastiaanWouters/verigo/ast"
	"github.com/SebastiaanWouters/verigo/evaluator"
	"github.com/SebastiaanWouters/verigo/evaluator_simple"
	"github.com/SebastiaanWouters/verigo/lexer"
	"github.com/SebastiaanWouters/verigo/object"
)

const PROMPT = ">> "

func opChanMonitor(c chan int) {
	for {
		//Gets called when an operation executes, contains the opcode (implement counting logic for the REPL here)
		<-c
	}
}
func rChanMonitor(c chan object.Result) {
	for {
		writeToDisk(<-c)
	}
}

func writeToDisk(res object.Result) {
	filename := "results.json"

	err := checkFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}

	data := []object.Result{}

	json.Unmarshal(file, &data)

	data = append(data, res)

	// Preparing the data to be marshalled and written.
	dataBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile(filename, dataBytes, 0644)
	if err != nil {
		fmt.Println(err)
	}

}

func checkFile(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	opChan := make(chan int)
	rChan := make(chan object.Result)
	go opChanMonitor(opChan)
	go rChanMonitor(rChan)

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

		evaluator.Eval(program, env, rChan, opChan)
	}
}

func Eval(input string, rChan chan object.Result, opChan chan int) {
	env := object.NewEnvironment()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	evaluator.Eval(program, env, rChan, opChan)

}

func EvalParsed(program *ast.Program, env *object.Environment, rChan chan object.Result, opChan chan int) {
	evaluator.Eval(program, env, rChan, opChan)
}

func Eval_Simple(input string) {
	env := object.NewEnvironment()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	evaluator_simple.Eval(program, env)

}

func EvalParsed_Simple(program *ast.Program, env *object.Environment) {
	evaluator_simple.Eval(program, env)
}

func Eval_Middle(input string, opCount *int) {
	env := object.NewEnvironment()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	evaluator_middle.Eval(program, env, opCount)

}

func EvalParsed_Middle(program *ast.Program, env *object.Environment, opCount *int) {
	evaluator_middle.Eval(program, env, opCount)
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
