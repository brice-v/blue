package repl

import (
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

// StartVmRepl start the read Vm print loop for the parser
func StartVmRepl() {
	startVmRepl(os.Stdin, os.Stdout, getUsername(), "", "")
}

// startVmRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output
func startVmRepl(in io.ReadCloser, out io.Writer, username, nodeName, address string) {
	rl := NewReadline(in, out, "VM", username)
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0, v.Help())
	}
	for i, v := range object.BuiltinobjsList {
		symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex, v.Builtin.Help())
	}
	_, err := fmt.Fprintln(out, "type .help for more information or help(OBJECT) for a specific object")
	if err != nil {
		log.Printf("Failed to write to repl output, error: %s", err.Error())
	}
	var filebuf bytes.Buffer
	replVarIndx := 1
	var c *compiler.Compiler = nil
	for {
		line := readLine(rl)
		if strings.HasPrefix(line, ".") {
			if strings.HasPrefix(line, ".exit") {
				_, err = io.WriteString(out, "\n")
				if err != nil {
					log.Printf("Failed to write to repl output, error: %s", err.Error())
				}
				break
			}
			// TODO: Need to be able to pass something here to store loaded file
			err := handleVmDotCommand(line, out, &filebuf, nil)
			if err != nil {
				_, errr := fmt.Fprintf(out, "repl command error: %s\n", err.Error())
				if errr != nil {
					log.Printf("Failed to write to repl output, error: %s", errr.Error())
				}
			}
			continue
		}

		l := lexer.New(line, "<repl>")
		p := parser.New(l)
		program := p.ParseProgram()
		if p.HasErrors() {
			p.PrintParserErrors(out)
			continue
		}
		if c == nil {
			c = compiler.NewWithStateAndCore(symbolTable, constants)
		}
		err := c.Compile(program)
		if err != nil {
			errToPrint, _, _ := strings.Cut(err.Error(), "\n"+consts.INTERNAL_ERROR_PATTERN)
			consts.ErrorPrinter(fmt.Sprintf("%s%s\n", consts.COMPILER_ERROR_PREFIX, errToPrint))
			c.PrintStackTrace()
			continue
		}
		bc := c.Bytecode()
		constants = bc.Constants
		v := vm.NewWithGlobalsStore(bc, globals)
		err = v.Run()
		if err == nil {
			replVar := fmt.Sprintf("_%d", replVarIndx)
			symbol := symbolTable.Define(replVar, true)
			globals[symbol.Index] = v.LastPoppedStackElem()
			replVarIndx++
			_, errr := fmt.Fprintf(out, "%s => %s\n", replVar, v.LastPoppedStackElem().Inspect())
			if errr != nil {
				log.Printf("Failed to write to repl output, error: %s", errr.Error())
			}
		} else {
			_, errr := fmt.Fprintf(out, "%s\n", err.Error())
			if errr != nil {
				log.Printf("Failed to write to repl output, error: %s", errr.Error())
			}
		}
		_, errr := fmt.Fprintf(&filebuf, "%s\n", line)
		if errr != nil {
			log.Printf("Failed to write to repl output, error: %s", errr.Error())
		}
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
			_, errr := fmt.Fprintf(out, "%+v\n", tok)
			if errr != nil {
				log.Printf("Failed to write to repl output, error: %s", errr.Error())
			}
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
		if p.HasErrors() {
			p.PrintParserErrors(out)
			continue
		}
		_, errr := fmt.Fprintf(out, "%s\n", program.String())
		if errr != nil {
			log.Printf("Failed to write to repl output, error: %s", errr.Error())
		}
	}
}

func NewReadline(in io.ReadCloser, out io.Writer, mode, username string) *readline.Instance {
	_, errr := fmt.Fprintf(out, "blue | v%s | REPL | MODE: %s | User: %s\n", consts.VERSION, mode, username)
	if errr != nil {
		log.Printf("Failed to write to repl output, error: %s", errr.Error())
	}
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
