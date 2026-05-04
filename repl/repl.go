package repl

import (
	"blue/compiler"
	"blue/consts"
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/token"
	"blue/vm"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/chzyer/readline"
)

// PROMPT is printed to the screen every time the user can type
const PROMPT = "> "

// StartLexerRepl starts the read eval print loop for the lexer
func StartLexerRepl() {
	startLexerRepl(os.Stdin, os.Stdout, getUsername())
}

// StartParserRepl start the read eval print loop for the parser
func StartParserRepl() {
	startParserRepl(os.Stdin, os.Stdout, getUsername())
}

// StartEvalRepl start the read eval print loop for the parser
func StartEvalRepl() {
	startEvalRepl(os.Stdin, os.Stdout, getUsername(), "", "")
}

// StartVmRepl start the read Vm print loop for the parser
func StartVmRepl() {
	startVmRepl(os.Stdin, os.Stdout, getUsername(), "", "")
}

// startEvalRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output
func startEvalRepl(in io.ReadCloser, out io.Writer, username, nodeName, address string) {
	rl := NewReadline(in, out, "EVAL", username)
	fmt.Fprintln(out, "type .help for more information or help(OBJECT) for a specific object")
	var filebuf bytes.Buffer
	replVarIndx := 1
	e := evaluator.NewNode(nodeName, address)
	for {
		line := readLine(rl)
		if strings.HasPrefix(line, ".") {
			if strings.HasPrefix(line, ".exit") {
				io.WriteString(out, "\n")
				break
			}
			err := handleDotCommand(line, out, &filebuf, e)
			if err != nil {
				fmt.Fprintf(out, "repl command error: %s\n", err.Error())
			}
			continue
		}

		l := lexer.New(line, "<repl>")
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			PrintParserErrors(out, p.Errors())
			continue
		}
		evaluated := e.Eval(program)

		if evaluated != nil {
			replVar := fmt.Sprintf("_%d", replVarIndx)
			e.ReplEnvAdd(replVar, evaluated)
			replVarIndx++
			fmt.Fprintf(out, "%s => %s\n", replVar, evaluated.Inspect())
		}
		fmt.Fprintf(&filebuf, "%s\n", line)
	}
}

// startEvalRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output
func startVmRepl(in io.ReadCloser, out io.Writer, username, nodeName, address string) {
	rl := NewReadline(in, out, "VM", username)
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0)
	}
	for i, v := range object.BuiltinobjsList {
		symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex)
	}
	fmt.Fprintln(out, "type .help for more information or help(OBJECT) for a specific object")
	var filebuf bytes.Buffer
	replVarIndx := 1
	for {
		line := readLine(rl)
		if strings.HasPrefix(line, ".") {
			if strings.HasPrefix(line, ".exit") {
				io.WriteString(out, "\n")
				break
			}
			// TODO: Need to be able to pass something here to store loaded file
			err := handleVmDotCommand(line, out, &filebuf, nil)
			if err != nil {
				fmt.Fprintf(out, "repl command error: %s\n", err.Error())
			}
			continue
		}

		l := lexer.New(line, "<repl>")
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			PrintParserErrors(out, p.Errors())
			continue
		}
		c := compiler.NewWithStateAndCore(symbolTable, constants)
		err := c.Compile(program)
		if err != nil {
			consts.ErrorPrinter(fmt.Sprintf("%s%s\n", consts.COMPILER_ERROR_PREFIX, err.Error()))
			c.PrintStackTrace()
			continue
		}
		bc := c.Bytecode()
		constants = bc.Constants
		v := vm.NewWithGlobalsStore(bc, globals)
		err = v.Run()
		if err == nil {
			replVar := fmt.Sprintf("_%d", replVarIndx)
			// TODO: Add var to environment
			replVarIndx++
			fmt.Fprintf(out, "%s => %s\n", replVar, v.LastPoppedStackElem().Inspect())
		} else {
			fmt.Fprintf(out, "%s\n", err.Error())
		}
		fmt.Fprintf(&filebuf, "%s\n", line)
	}
}

// startLexerRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output
func startLexerRepl(in io.ReadCloser, out io.Writer, username string) {
	rl := NewReadline(in, out, "LEX", username)
	for {
		line := readLine(rl)
		l := lexer.New(line, "<repl>")
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
}

// PrintParserErrors prints the parser errors to the output
func PrintParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		splitMsg := strings.Split(msg, "\n")
		firstPart := consts.PARSER_ERROR_PREFIX + splitMsg[0] + "\n"
		consts.ErrorPrinter(firstPart)
		for i, s := range splitMsg {
			if i == 0 {
				continue
			}
			fmt.Fprintf(out, "%s\n", s)
		}
	}
}

// startParserRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output
func startParserRepl(in io.ReadCloser, out io.Writer, username string) {
	rl := NewReadline(in, out, "PARSE", username)
	for {
		line := readLine(rl)
		l := lexer.New(line, "<repl>")
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			PrintParserErrors(out, p.Errors())
			continue
		}
		fmt.Fprintf(out, "%s\n", program.String())
	}
}

func NewReadline(in io.ReadCloser, out io.Writer, mode, username string) *readline.Instance {
	fmt.Fprintf(out, "blue | v%s | REPL | MODE: %s | User: %s\n", consts.VERSION, mode, username)
	rl, err := readline.NewEx(&readline.Config{Stdin: in, Stdout: out, Prompt: PROMPT})
	if err != nil {
		consts.ErrorPrinter("Failed to instantiate readline. error: %s\n", err.Error())
		os.Exit(1)
	}
	return rl
}

func getUsername() string {
	user, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current username, proceeding with none")
		return ""
	}
	return user.Username
}

func readLine(rl *readline.Instance) string {
	line, err := rl.Readline()
	if err != nil {
		if err.Error() == "Interrupt" || err.Error() == "EOF" {
			println(err.Error())
			os.Exit(0)
		}
		consts.ErrorPrinter("Failed to read line: Unexpected Error: %s\n", err.Error())
		os.Exit(1)
	}
	return line
}
