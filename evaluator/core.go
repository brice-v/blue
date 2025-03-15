package evaluator

import (
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"bytes"
	"fmt"
	"os"
)

var coreEnv *object.Environment2 = nil

func (e *Evaluator) AddCoreLibToEnv() *object.Environment2 {
	if coreEnv == nil {
		l := lexer.New(lib.CoreFile, consts.CORE_FILE_PATH)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				consts.ErrorPrinter("ParserError in core.b: %s\n", msg)
			}
			os.Exit(1)
		}
		result := e.Eval(program)
		if isError(result) {
			errorObj := result.(*object.Error)
			var buf bytes.Buffer
			buf.WriteString(errorObj.Message)
			buf.WriteByte('\n')
			for e.ErrorTokens.Len() > 0 {
				buf.WriteString(lexer.GetErrorLineMessage(e.ErrorTokens.PopBack()))
				buf.WriteByte('\n')
			}
			fmt.Printf("%s%s", consts.EVAL_ERROR_PREFIX, buf.String())
			os.Exit(1)
		}
		coreEnv = e.env.Clone()
	}
	return coreEnv
}
