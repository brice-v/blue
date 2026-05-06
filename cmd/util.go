package cmd

import (
	"blue/blueutil"
	"blue/code"
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

// evalFileOrString evaluates the given file or string if isFpath is false
func evalFileOrString(inputOrFpath string, isFpath, noExec bool) {
	var l *lexer.Lexer
	if isFpath {
		data, err := os.ReadFile(inputOrFpath)
		if err != nil {
			consts.ErrorPrinter("`evalFile` error trying to read file `%s`. error: %s\n", inputOrFpath, err.Error())
			os.Exit(1)
		}
		l = lexer.New(string(data), inputOrFpath)
	} else {
		l = lexer.New(inputOrFpath, "<stdin>")
	}
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
	object.NoExec = noExec
	e := evaluator.New()
	if isFpath {
		e.CurrentFile = filepath.Clean(inputOrFpath)
		e.EvalBasePath = filepath.Dir(inputOrFpath)
	}
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

func instantiateCompiler(inputOrFpath string, isFpath bool) *compiler.Compiler {
	var l *lexer.Lexer
	if isFpath {
		data, err := os.ReadFile(inputOrFpath)
		if err != nil {
			consts.ErrorPrinter("`vm` error trying to read file `%s`. error: %s\n", inputOrFpath, err.Error())
			os.Exit(1)
		}
		l = lexer.New(string(data), inputOrFpath)
	} else {
		l = lexer.New(inputOrFpath, "<stdin>")
	}

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
	constants := object.NewObjectConstants()
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0)
	}
	for i, v := range object.BuiltinobjsList {
		symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex)
	}
	c := compiler.NewWithStateAndCore(symbolTable, constants)
	if err := c.Compile(program); err != nil {
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, err.Error())
		c.PrintStackTrace()
		os.Exit(1)
	}
	return c
}

func compileFileOrString(inputOrFpath string, isFpath bool) {
	c := instantiateCompiler(inputOrFpath, isFpath)
	offset := 0
	for i, ins := range c.Bytecode().Instructions {
		if ins == byte(code.OpCoreCompiled) {
			offset = i
		}
	}
	fmt.Print(blueutil.BytecodeDebugStringWithOffset(offset, c.Bytecode().Instructions[offset:], c.Bytecode().Constants))
	os.Exit(0)
}

func vmFileOrString(inputOrFpath string, isFpath, noExec bool) {
	c := instantiateCompiler(inputOrFpath, isFpath)
	globals := make([]object.Object, vm.GlobalsSize)
	bc := c.Bytecode()
	v := vm.NewWithGlobalsStore(bc, globals)
	object.NoExec = noExec
	if err := v.Run(); err != nil {
		consts.ErrorPrinter("%s%s\n", consts.VM_ERROR_PREFIX, err.Error())
		if v.TokensForErrorTrace != nil {
			for _, tok := range v.TokensForErrorTrace {
				fmt.Println(lexer.GetErrorLineMessage(*tok))
			}
		}
		os.Exit(1)
	}
	val := v.LastPoppedStackElem()
	if val.Type() == object.ERROR_OBJ {
		errorObj := val.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
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
