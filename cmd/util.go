package cmd

import (
	"blue/consts"
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"blue/token"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var out = os.Stdout

// isFile is a helper function to check if the fpath given
// exists and if not return false
func isFile(fpath string) bool {
	info, err := os.Stat(fpath)
	if err == nil {
		return !info.IsDir()
	}
	return false
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
		log.Fatalf("`lexFile` error trying to read file `%s`. error: %s", fpath, err.Error())
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
		log.Fatalf("`parseFile` error trying to read file `%s`. error: %s", fpath, err.Error())
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
func evalFile(fpath string) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("`evalFile` error trying to read file `%s`. error: %s", fpath, err.Error())
	}

	l := lexer.New(string(data), fpath)

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
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
		out.WriteString(fmt.Sprintf("%s%s", consts.EVAL_ERROR_PREFIX, buf.String()))
		os.Exit(1)
	}
	// NOTE: This could be used for debugging programs return values
	// if evaluated != nil {
	// 	os.Stdout.WriteString(evaluated.Inspect() + "\n")
	// }
}

// evalString evaluates the given string
func evalString(strToEval string) {
	l := lexer.New(strToEval, "<stdin>")

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
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
		out.WriteString(fmt.Sprintf("%s%s", consts.EVAL_ERROR_PREFIX, buf.String()))
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
			log.Fatalf("`doc` error trying to read file `%s`. error: %s", name, err.Error())
		}
		l := lexer.New(string(fdata), name)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				fmt.Printf("ParserError in `%s` module: %s\n", name, msg)
			}
			os.Exit(1)
		}
		e.Eval(program)
		pubFunHelpStr := e.GetPublicFunctionHelpString()
		return evaluator.CreateHelpStringFromProgramTokens(name, program.HelpStrTokens, pubFunHelpStr) + "\n"
	}
	return ""
}
