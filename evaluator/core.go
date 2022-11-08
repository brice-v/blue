package evaluator

import (
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"bytes"
	"fmt"
	"os"

	_ "embed"
)

//go:embed core/core.b
var coreFile string

func (e *Evaluator) AddCoreLibToEnv() {
	l := lexer.New(coreFile, "<embed: core/core.b>")

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
		errorObj := result.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		for e.ErrorTokens.Len() > 0 {
			buf.WriteString(l.GetErrorLineMessage(e.ErrorTokens.PopBack()))
			buf.WriteByte('\n')
		}
		fmt.Printf("EvaluatorError: %s", buf.String())
		os.Exit(1)
	}
}
