package cmd

import (
	"blue/compiler"
	"blue/consts"
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"blue/token"
	"blue/vm"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var out = os.Stdout

// isFile is a helper function to check if the fpath given
// exists and if not return false
func isFile(fpath string) bool {
	info, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// isDir is a helper function to check if the dirPath given
// exists and if its a dir otherwise false
func isDir(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err == nil {
		return info.IsDir()
	}
	return false
}

// lexFile tokenizes and lexically analyzes the given file
func lexFile(fpath string) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		consts.ErrorPrinter("`lexFile` error trying to read file `%s`. error: %s\n", fpath, err.Error())
		os.Exit(1)
	}

	l := lexer.New(string(data), fpath)

	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		fmt.Printf("%+v\n", tok)
	}
}

// parseFile parses the given file
func parseFile(fpath string) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		consts.ErrorPrinter("`parseFile` error trying to read file `%s`. error: %s\n", fpath, err.Error())
		os.Exit(1)
	}

	l := lexer.New(string(data), fpath)

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}

	io.WriteString(out, program.String())
	io.WriteString(out, "\n")
}

// evalFile evaluates the given file
func evalFile(fpath string, noExec bool) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		consts.ErrorPrinter("`evalFile` error trying to read file `%s`. error: %s\n", fpath, err.Error())
		os.Exit(1)
	}

	l := lexer.New(string(data), fpath)

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
	evaluator.NoExec = noExec
	e := evaluator.New()
	e.CurrentFile = filepath.Clean(fpath)
	e.EvalBasePath = filepath.Dir(fpath)
	val := e.Eval(program)
	if val.Type() == object.ERROR_OBJ {
		errorObj := val.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		for e.ErrorTokens.Len() > 0 {
			buf.WriteString(lexer.GetErrorLineMessage(e.ErrorTokens.PopBack()))
			buf.WriteByte('\n')
		}
		msg := fmt.Sprintf("%s%s", consts.EVAL_ERROR_PREFIX, buf.String())
		splitMsg := strings.Split(msg, "\n")
		for i, s := range splitMsg {
			if i == 0 {
				consts.ErrorPrinter(s + "\n")
				continue
			}
			delimeter := ""
			if i != len(splitMsg)-1 {
				delimeter = "\n"
			}
			fmt.Fprintf(out, "%s%s", s, delimeter)
		}
		os.Exit(1)
	}
	// NOTE: This could be used for debugging programs return values
	// if evaluated != nil {
	// 	os.Stdout.WriteString(evaluated.Inspect() + "\n")
	// }
}

func vmFile(fpath string, noExec bool) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		consts.ErrorPrinter("`vm` error trying to read file `%s`. error: %s\n", fpath, err.Error())
		os.Exit(1)
	}

	l := lexer.New(string(data), fpath)

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	compiled := compiler.NewWithState(symbolTable, constants)
	err = compiled.Compile(program)
	if err != nil {
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, err.Error())
		os.Exit(1)
	}
	bc := compiled.Bytecode()
	v := vm.NewWithGlobalsStore(bc, globals)
	err = v.Run()
	if err != nil {
		consts.ErrorPrinter("`%s%s\n", consts.VM_ERROR_PREFIX, err.Error())
		os.Exit(1)
	}
	val := v.LastPoppedStackElem()
	if val.Type() == object.ERROR_OBJ {
		errorObj := val.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		// for e.ErrorTokens.Len() > 0 {
		// 	buf.WriteString(lexer.GetErrorLineMessage(e.ErrorTokens.PopBack()))
		// 	buf.WriteByte('\n')
		// }
		msg := fmt.Sprintf("%s%s", consts.VM_ERROR_PREFIX, buf.String())
		splitMsg := strings.Split(msg, "\n")
		for i, s := range splitMsg {
			if i == 0 {
				consts.ErrorPrinter(s + "\n")
				continue
			}
			delimeter := ""
			if i != len(splitMsg)-1 {
				delimeter = "\n"
			}
			fmt.Fprintf(out, "%s%s", s, delimeter)
		}
		os.Exit(1)
	}
}

// evalString evaluates the given string
func evalString(strToEval string, noExec bool) {
	l := lexer.New(strToEval, "<stdin>")

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
	evaluator.NoExec = noExec
	e := evaluator.New()
	val := e.Eval(program)
	if val.Type() == object.ERROR_OBJ {
		errorObj := val.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		for e.ErrorTokens.Len() > 0 {
			buf.WriteString(lexer.GetErrorLineMessage(e.ErrorTokens.PopBack()))
			buf.WriteByte('\n')
		}
		msg := fmt.Sprintf("%s%s", consts.EVAL_ERROR_PREFIX, buf.String())
		splitMsg := strings.Split(msg, "\n")
		for i, s := range splitMsg {
			if i == 0 {
				consts.ErrorPrinter(s + "\n")
				continue
			}
			delimeter := ""
			if i != len(splitMsg)-1 {
				delimeter = "\n"
			}
			fmt.Fprintf(out, "%s%s", s, delimeter)
		}
		os.Exit(1)
	}
	// NOTE: This could be used for debugging programs return values
	// if evaluated != nil {
	// 	os.Stdout.WriteString(evaluated.Inspect() + "\n")
	// }
}

func getDocStringFor(name string) string {
	e := evaluator.New()
	if name == "std" {
		// Get all std modules public function help strings
		return e.GetAllStdPublicFunctionHelpStrings()
	}
	if e.IsStd(name) {
		// Get module's public function help string
		return e.GetStdModPublicFunctionHelpString(name)
	}
	if isFile(name) {
		fdata, err := os.ReadFile(name)
		if err != nil {
			consts.ErrorPrinter("`doc` error trying to read file `%s`. error: %s\n", name, err.Error())
			os.Exit(1)
		}
		l := lexer.New(string(fdata), name)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				splitMsg := strings.Split(msg, "\n")
				firstPart := fmt.Sprintf("%smodule `%s`: %s\n", consts.PARSER_ERROR_PREFIX, name, splitMsg[0])
				consts.ErrorPrinter(firstPart)
				for i, s := range splitMsg {
					if i == 0 {
						continue
					}
					fmt.Println(s)
				}
			}
			os.Exit(1)
		}
		e.Eval(program)
		pubFunHelpStr := e.GetPublicFunctionHelpString()
		return evaluator.CreateHelpStringFromProgramTokens(name, program.HelpStrTokens, pubFunHelpStr) + "\n"
	}
	return ""
}
