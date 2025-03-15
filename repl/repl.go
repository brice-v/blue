package repl

import (
	"blue/consts"
	"blue/evaluator"
	"blue/lexer"
	"blue/parser"
	"blue/token"
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
	user, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current username, proceeding with none")
		startLexerRepl(os.Stdin, os.Stdout, "")
		os.Exit(0)
	}
	startLexerRepl(os.Stdin, os.Stdout, user.Username)
	os.Exit(0)
}

// StartParserRepl start the read eval print loop for the parser
func StartParserRepl() {
	user, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current username, proceeding with none")
		startParserRepl(os.Stdin, os.Stdout, "")
		os.Exit(0)
	}
	startParserRepl(os.Stdin, os.Stdout, user.Username)
	os.Exit(0)
}

// StartEvalRepl start the read eval print loop for the parser
func StartEvalRepl() {
	user, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current username, proceeding with none")
		startEvalRepl(os.Stdin, os.Stdout, "", "", "")
		os.Exit(0)
	}
	startEvalRepl(os.Stdin, os.Stdout, user.Username, "", "")
	os.Exit(0)
}

func StartEvalReplWithNodeName(nodeName, address string) {
	user, err := user.Current()
	if err != nil {
		fmt.Println("Unable to get current username, proceeding with none")
		startEvalRepl(os.Stdin, os.Stdout, "", nodeName, address)
		os.Exit(0)
	}
	startEvalRepl(os.Stdin, os.Stdout, user.Username, nodeName, address)
	os.Exit(0)
}

// startEvalRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output
func startEvalRepl(in io.Reader, out io.Writer, username, nodeName, address string) {
	e := evaluator.NewNode(nodeName, address)
	header := fmt.Sprintf("blue | v%s | REPL | MODE: EVAL | User: %s", consts.VERSION, username)
	rl, err := readline.New(PROMPT)
	if err != nil {
		consts.ErrorPrinter("Failed to instantiate readline| Error: %s", err)
		os.Exit(1)
	}
	consts.InfoPrinter(header + "\n")
	fmt.Println("type .help for more information or help(OBJECT) for a specific object")
	var filebuf bytes.Buffer
	replVarIndx := 1
	for {
		line, err := rl.Readline()
		if err != nil {
			if err.Error() == "Interrupt" || err.Error() == "EOF" {
				println(err.Error())
				os.Exit(0)
			}
			consts.ErrorPrinter("Failed to read line: Unexpected Error: %s", err.Error())
			os.Exit(1)
			break
		}

		if strings.HasPrefix(line, ".") {
			if strings.HasPrefix(line, ".exit") {
				io.WriteString(out, "\n")
				break
			}
			err := handleDotCommand(line, out, &filebuf, e)
			if err != nil {
				io.WriteString(out, "repl command error: ")
				io.WriteString(out, err.Error())
				io.WriteString(out, "\n")
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
			io.WriteString(out, replVar)
			io.WriteString(out, " => ")
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
		filebuf.WriteString(line)
		filebuf.WriteByte('\n')
	}
}

// startLexerRepl is the entry point of the repl with an io.Reader as
// an input and io.Writer as an output
func startLexerRepl(in io.Reader, out io.Writer, username string) {
	header := fmt.Sprintf("blue | v%s | REPL | MODE: LEXER | User: %s", consts.VERSION, username)
	rl, err := readline.New(PROMPT)
	if err != nil {
		consts.ErrorPrinter("Failed to instantiate readline| Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(header)
	for {
		line, err := rl.Readline()
		if err != nil {
			if err.Error() == "Interrupt" || err.Error() == "EOF" {
				println(err.Error())
				os.Exit(0)
			}
			consts.ErrorPrinter("Failed to read line: Unexpected Error: %s\n", err.Error())
			os.Exit(1)
			break
		}

		l := lexer.New(line, "<repl>")

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Printf("%+v\n", tok)
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
func startParserRepl(in io.Reader, out io.Writer, username string) {
	header := fmt.Sprintf("blue | v%s | REPL | MODE: PARSER | User: %s", consts.VERSION, username)
	rl, err := readline.New(PROMPT)
	if err != nil {
		consts.ErrorPrinter("Failed to instantiate readline| Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(header)
	for {
		line, err := rl.Readline()
		if err != nil {
			if err.Error() == "Interrupt" || err.Error() == "EOF" {
				println(err.Error())
				os.Exit(0)
			}
			consts.ErrorPrinter("Failed to read line: Unexpected Error: %s\n", err.Error())
			os.Exit(1)
			break
		}

		l := lexer.New(line, "<repl>")

		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			PrintParserErrors(out, p.Errors())
			continue
		}

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}
