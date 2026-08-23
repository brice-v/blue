package cmd

import (
	"blue/ast"
	"blue/blueutil"
	"blue/code"
	"blue/compiler"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/token"
	"blue/vm"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// out is where normal program and command output is written
var out = os.Stdout

// isFile checks whether fpath exists and is not a directory.
func isFile(fpath string) bool {
	info, err := os.Stat(fpath)
	return !os.IsNotExist(err) && !info.IsDir()
}

// lexFile tokenizes and lexically analyzes the given file
func lexFile(fpath string) {
	var data []byte
	var err error
	fname := fpath
	if fpath == STDIN_ARG {
		data, err = io.ReadAll(os.Stdin)
		fname = STDIN_NAME
	} else {
		data, err = os.ReadFile(fpath)
	}
	if err != nil {
		consts.ErrorPrinter("`lexFile` error trying to read file `%s`. error: %s\n", fpath, err.Error())
		os.Exit(1)
	}

	l := lexer.New(string(data), fname)

	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		fmt.Printf("%+v\n", tok)
	}
}

// parseFile parses the given file
func parseFile(fpath string, allErrors bool) {
	program := lexAndParse(fpath, true, allErrors)
	_, err := io.WriteString(out, program.String())
	if err != nil {
		log.Printf("Failed to write string to out parameter, error: %s", err.Error())
	}
	_, err = io.WriteString(out, "\n")
	if err != nil {
		log.Printf("Failed to write string to out parameter, error: %s", err.Error())
	}
}

// STDIN_ARG is the conventional argument that means read the program from STDIN
const STDIN_ARG = "-"

// STDIN_NAME is the name reported in error traces for programs read from STDIN
const STDIN_NAME = "<stdin>"

// stdinIsTerminal reports whether stdin is attached to a terminal as
// opposed to being piped or redirected
func stdinIsTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func lexAndParse(inputOrFpath string, isFpath bool, allErrors bool) *ast.Program {
	var l *lexer.Lexer
	if inputOrFpath == STDIN_ARG {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			consts.ErrorPrinter("error trying to read from stdin. error: %s\n", err.Error())
			os.Exit(1)
		}
		l = lexer.New(string(data), STDIN_NAME)
	} else if isFpath {
		data, err := os.ReadFile(inputOrFpath)
		if err != nil {
			consts.ErrorPrinter("error trying to read file `%s`. error: %s\n", inputOrFpath, err.Error())
			os.Exit(1)
		}
		l = lexer.New(string(data), inputOrFpath)
	} else {
		l = lexer.New(inputOrFpath, "<stdin>")
	}

	var p *parser.Parser
	if allErrors {
		p = parser.New(l)
	} else {
		p = parser.NewWithStopAfterFirst(l)
	}
	program := p.ParseProgram()
	if p.HasErrors() {
		p.PrintParserErrors(os.Stderr)
		os.Exit(1)
	}
	return program
}

func newCompiler(isFpath bool, fpath string) *compiler.Compiler {
	constants := object.NewObjectConstants()
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0, v.Help())
	}
	for i, v := range object.BuiltinobjsList {
		symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex, v.Builtin.Help())
	}
	c := compiler.NewWithStateAndCore(symbolTable, constants)
	if isFpath {
		c.CompilerBasePath = filepath.Dir(fpath)
	}
	return c
}

func compileProgram(c *compiler.Compiler, program *ast.Program) {
	if err := c.Compile(program); err != nil {
		errToPrint, _, _ := strings.Cut(err.Error(), "\n"+consts.INTERNAL_ERROR_PATTERN)
		consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, errToPrint)
		c.PrintStackTrace()
		os.Exit(1)
	}
}

func instantiateCompiler(inputOrFpath string, isFpath bool, allErrors bool) *compiler.Compiler {
	program := lexAndParse(inputOrFpath, isFpath, allErrors)
	c := newCompiler(isFpath, inputOrFpath)
	compileProgram(c, program)
	return c
}

func instantiateCompilerForDoc(fpath string) string {
	modName := strings.ReplaceAll(filepath.Base(fpath), ".b", "")
	program := lexAndParse(fpath, true, false)
	c := newCompiler(true, fpath)
	c.SetDocModName(modName)
	compileProgram(c, program)
	pubFunHelpStr := c.GetDocOrderedPublicFunctionHelpString(modName)
	return object.CreateHelpStringFromProgramTokens(modName, program.HelpStrTokens, pubFunHelpStr) + "\n"
}

func compileFileOrString(inputOrFpath string, isFpath bool, allErrors bool) {
	c := instantiateCompiler(inputOrFpath, isFpath, allErrors)
	offset := 0
	for i, ins := range c.Bytecode().Instructions {
		if ins == byte(code.OpCoreCompiled) {
			offset = i
		}
	}
	fmt.Print(blueutil.BytecodeDebugStringWithOffset(offset, c.Bytecode().Instructions[offset:], c.Bytecode().Constants))
	os.Exit(0)
}

func vmFileOrString(inputOrFpath string, isFpath, noExec, allErrors, printResult bool) {
	c := instantiateCompiler(inputOrFpath, isFpath, allErrors)
	globals := make([]object.Object, vm.GlobalsSize)
	bc := c.Bytecode()
	v := vm.NewWithGlobalsStore(bc, globals)
	object.NoExec = noExec
	if err := v.Run(); err != nil {
		if v.TokensForErrorTrace == nil {
			consts.ErrorPrinter("%s%s\n", consts.VM_ERROR_PREFIX, err.Error())
		} else {
			for i, tok := range v.TokensForErrorTrace {
				errorLine := lexer.GetErrorLineMessage(*tok)
				fullMsg := fmt.Sprintf("%s\n%s", err.Error(), errorLine)
				blueutil.PrintCustomError(os.Stderr, consts.VM_ERROR_PREFIX, fullMsg, tok.LineNumber, i == 0)
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
				consts.ErrorPrinter("%s\n", s)
				continue
			}
			delimeter := ""
			if i != len(splitMsg)-1 {
				delimeter = "\n"
			}
			_, errr := fmt.Fprintf(os.Stderr, "%s%s", s, delimeter)
			if errr != nil {
				log.Printf("Failed to write to output, error: %s", errr.Error())
			}
		}
		os.Exit(1)
	}
	if printResult {
		_, err := fmt.Fprintf(out, "%s\n", val.Inspect())
		if err != nil {
			log.Printf("Failed to write to output, error: %s", err.Error())
		}
	}
}

func getBuiltinHelpIfExists(name string) string {
	var out bytes.Buffer
	found := false
	// Look through modules
	for _, builtins := range object.AllBuiltins {
		if builtins.Name == name {
			found = true
			fmt.Fprintf(&out, "MODULE: %s\n", name)
			for _, b := range builtins.Builtins {
				fmt.Fprintf(&out, "%s\n", b.HelpStr)
			}
		}
	}
	// Look through builtins individually
	if !found {
		for _, builtins := range object.AllBuiltins {
			for _, b := range builtins.Builtins {
				if b.Name == name || b.Name[1:] == name {
					fmt.Fprintf(&out, "%s", b.HelpStr)
				}
			}
		}
	}
	return out.String()
}

func getDocStringFor(name string) string {
	builtinHelpStr := getBuiltinHelpIfExists(name)
	if builtinHelpStr != "" {
		return builtinHelpStr
	}
	if name == "std" {
		mods := compiler.StdModuleNames()
		sort.Strings(mods)
		var out bytes.Buffer
		for i, mod := range mods {
			c := compiler.NewFromCore()
			out.WriteString(c.GetStdModuleDocString(mod))
			if i != len(mods)-1 {
				out.WriteByte('\n')
			}
		}
		return out.String()
	}
	if compiler.IsStd(name) {
		c := compiler.NewFromCore()
		return c.GetStdModuleDocString(name)
	}
	if isFile(name) {
		return instantiateCompilerForDoc(name)
	}
	return ""
}
