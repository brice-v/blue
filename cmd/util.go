package cmd

import (
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
	object.NoExec = noExec
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

func vmFile(fpath string, noExec bool, compile bool) {
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
	constants := object.NewObjectConstants()
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0)
	}
	c := compiler.NewWithStateAndCore(symbolTable, constants)
	err = c.Compile(program)
	if err != nil {
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, err.Error())
		c.PrintStackTrace()
		os.Exit(1)
	}
	if compile {
		offset := 0
		for i, ins := range c.Bytecode().Instructions {
			if ins == byte(code.OpCoreCompiled) {
				offset = i
			}
		}
		fmt.Print(BytecodeDebugStringWithOffset(offset, c.Bytecode().Instructions[offset:], c.Bytecode().Constants))
		os.Exit(0)
	}
	bc := c.Bytecode()
	v := vm.NewWithGlobalsStore(bc, c.Tokens, globals)
	err = v.Run()
	if err != nil {
		consts.ErrorPrinter("%s%s\n", consts.VM_ERROR_PREFIX, err.Error())
		if v.TokensForErrorTrace != nil {
			for _, tok := range v.TokensForErrorTrace {
				fmt.Println(lexer.GetErrorLineMessage(tok))
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
	object.NoExec = noExec
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

func BytecodeDebugStringWithOffset(offset int, ins code.Instructions, constants []object.Object) string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := code.Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := code.ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", offset+i, fmtInstruction(def, operands, constants))
		i += 1 + read
	}
	return out.String()
}

func BytecodeDebugString(ins code.Instructions, constants []object.Object) string {
	return BytecodeDebugStringWithOffset(0, ins, constants)
}

func fmtInstruction(def *code.Definition, operands []int, constants []object.Object) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		lastPart := ""
		if def.Name == "OpConstant" {
			lastPart = fmt.Sprintf(" (%s)", constants[operands[0]].Inspect())
		}
		return fmt.Sprintf("%s %d%s", def.Name, operands[0], lastPart)
	case 2:
		lastPart := ""
		switch def.Name {
		case "OpGetBuiltin":
			lastPart = fmt.Sprintf(" (%s)", object.AllBuiltins[operands[0]].Builtins[operands[1]].Name)
		case "OpClosure":
			cf := constants[operands[0]].(*object.CompiledFunction)
			lastPart = fmt.Sprintf("\n\t%s", strings.ReplaceAll(BytecodeDebugString(cf.Instructions, constants), "\n", "\n\t"))
			lastPart = strings.TrimSuffix(lastPart, "\n\t")
		}
		return fmt.Sprintf("%s %d %d%s", def.Name, operands[0], operands[1], lastPart)
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}
