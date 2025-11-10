package compiler

import (
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/parser"
	"os"
)

func (c *Compiler) compileCore() {
	if !c.coreCompiled {
		l := lexer.New(lib.CoreFile, consts.CORE_FILE_PATH)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				consts.ErrorPrinter("ParserError in core.b: %s\n", msg)
			}
			os.Exit(1)
		}
		err := c.Compile(program)
		if err != nil {
			consts.ErrorPrinter("Failed to compile core.b: %s\n", err.Error())
		}
		c.coreCompiled = true
	}
}
