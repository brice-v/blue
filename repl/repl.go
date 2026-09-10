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

// StartLexerRepl starts the read eval print loop for the lexer. It returns
// nil when the session ends (exit command, interrupt or EOF) and an error
// when readline fails to start.
func StartLexerRepl() error {
	return startLexerRepl(os.Stdin, os.Stdout, getUsername())
}

// StartParserRepl starts the read eval print loop for the parser. It returns
// nil when the session ends (interrupt or EOF) and an error when readline
// fails to start.
func StartParserRepl() error {
	return startParserRepl(os.Stdin, os.Stdout, getUsername())
}

// StartVmRepl starts the read Vm print loop. It returns nil when the session
// ends (exit command, interrupt or EOF) and an error when readline fails to
// start.
func StartVmRepl() error {
	return startVmRepl(os.Stdin, os.Stdout, getUsername(), "", "")
}

// startVmRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output. It returns nil when the session ends
// (exit command, interrupt or EOF) and an error when readline fails to start.
func startVmRepl(in io.ReadCloser, out io.Writer, username, nodeName, address string) error {
	rl, err := NewReadline(in, out, "VM", username)
	if err != nil {
		return err
	}
	constants := object.NewObjectConstants()
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0, v.Help())
	}
	for i, v := range object.BuiltinobjsList {
		symbolTable.DefineBuiltin(i, v.Name, object.BuiltinobjsModuleIndex, v.Builtin.Help())
	}
	_, err = fmt.Fprintln(out, "type .help for more information or help(OBJECT) for a specific object")
	if err != nil {
		log.Printf("Failed to write to repl output, error: %s", err.Error())
	}
	var filebuf bytes.Buffer
	replVarIndx := 1
	// replGlobalsHwm tracks one past the highest global index ever written
	// in this session (by any line's vm or by the repl result vars), so new
	// per-line vms seed their spawn snapshots correctly.
	replGlobalsHwm := 0
	var c *compiler.Compiler = nil
	for {
		line, done, err := readLine(rl)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
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
		err = c.Compile(program)
		if err != nil {
			errToPrint, _, _ := strings.Cut(err.Error(), "\n"+consts.INTERNAL_ERROR_PATTERN)
			consts.ErrorPrinter("%s%s\n", consts.COMPILER_ERROR_PREFIX, errToPrint)
			c.PrintStackTrace()
			continue
		}
		bc := c.Bytecode()
		constants = bc.Constants
		v := vm.NewWithGlobalsStore(bc, globals)
		// Globals written by earlier lines (vm sets and repl result vars)
		// must be visible to spawn-time snapshots taken by this vm.
		v.SetGlobalsHighWater(replGlobalsHwm)
		err = v.Run()
		if err == nil {
			replVar := fmt.Sprintf("_%d", replVarIndx)
			symbol := symbolTable.Define(replVar, true)
			globals[symbol.Index] = v.LastPoppedStackElem()
			// Written outside the vm loop; keep future spawn snapshots
			// correct. The result var index and anything this line's vm
			// wrote both raise the watermark.
			replGlobalsHwm = symbol.Index + 1
			if hw := v.GlobalsHighWater(); hw > replGlobalsHwm {
				replGlobalsHwm = hw
			}
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
	return nil
}

// startLexerRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output. It returns nil when the session ends
// (interrupt or EOF) and an error when readline fails to start.
func startLexerRepl(in io.ReadCloser, out io.Writer, username string) error {
	rl, err := NewReadline(in, out, "LEX", username)
	if err != nil {
		return err
	}
	done := false
	for !done {
		var line string
		line, done, err = readLine(rl)
		if err != nil {
			return err
		}
		l := lexer.New(line, "<repl>")
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			_, errr := fmt.Fprintf(out, "%+v\n", tok)
			if errr != nil {
				log.Printf("Failed to write to repl output, error: %s", errr.Error())
			}
		}
	}
	return nil
}

// startParserRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output. It returns nil when the session ends
// (interrupt or EOF) and an error when readline fails to start.
func startParserRepl(in io.ReadCloser, out io.Writer, username string) error {
	rl, err := NewReadline(in, out, "PARSE", username)
	if err != nil {
		return err
	}
	done := false
	for !done {
		var line string
		line, done, err = readLine(rl)
		if err != nil {
			return err
		}
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
	return nil
}

// NewReadline instantiates a readline session for the given mode. It returns
// an error when the underlying terminal cannot be attached.
func NewReadline(in io.ReadCloser, out io.Writer, mode, username string) (*readline.Instance, error) {
	_, errr := fmt.Fprintf(out, "blue | v%s | REPL | MODE: %s | User: %s\n", consts.VERSION, mode, username)
	if errr != nil {
		log.Printf("Failed to write to repl output, error: %s", errr.Error())
	}
	rl, err := readline.NewEx(&readline.Config{Stdin: in, Stdout: out, Prompt: PROMPT})
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate readline. error: %s", err.Error())
	}
	return rl, nil
}

func getUsername() string {
	user, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current username, proceeding with none")
		return ""
	}
	return user.Username
}

// readLine reads one line from the session. It returns done=true when the
// user interrupted the session or the input hit EOF, and an error for any
// other readline failure.
func readLine(rl *readline.Instance) (line string, done bool, err error) {
	line, err = rl.Readline()
	if err != nil {
		if err.Error() == "Interrupt" || err.Error() == "EOF" {
			println(err.Error())
			return "", true, nil
		}
		consts.ErrorPrinter("Failed to read line: Unexpected Error: %s\n", err.Error())
		return "", false, err
	}
	return line, false, nil
}
