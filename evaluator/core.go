package evaluator

import (
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"fmt"
	"os"

	_ "embed"
)

//go:embed core/core.b
var coreFile string

func (e *Evaluator) AddCoreLibToEnv() {
	l := lexer.New(coreFile)

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("ParserError in core.b: %s\n", msg)
		}
		os.Exit(1)
	}
	result := e.Eval(program)
	if result.Type() == object.ERROR_OBJ {
		err := result.(*object.Error).Message
		fmt.Printf("EvaluatorError in core.b: %s\n", err)
		os.Exit(1)
	}
}