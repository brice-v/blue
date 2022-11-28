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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// bundleFile takes the given file as an entry point
// and bundles the interpreter with the code into a go executable
func bundleFile(fpath string, isDebug bool) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("`bundleFile` error trying to read file `%s`. error: %s", fpath, err.Error())
	}

	d := string(data)
	if isDebug {
		fmt.Println("File Name: '" + fpath + "', Data: ")
		fmt.Printf("`%s`\n\n", d)
	}

	header := `package main

import (
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

var out = os.Stderr

//go:embed **/*.b *.b
var files embed.FS
`

	entryPointPath := fmt.Sprintf("const entryPointPath = `%s`\n", fpath)
	mainFunc := `func main() {
	entryPoint, err := files.ReadFile(entryPointPath)
	if err != nil {
		out.WriteString("Failed to read EntryPoint File '" + entryPointPath + "'\n")
		os.Exit(1)
	}
	input := string(entryPoint)
	evaluator.IsEmbed = true
	evaluator.Files = files
	l := lexer.New(input, "<embed: "+entryPointPath+">")
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(out, p.Errors())
		os.Exit(1)
	}
	evaluator := evaluator.New()
	evaluator.CurrentFile = "<embed>"
	evaluator.EvalBasePath = filepath.Dir(".")
	val := evaluator.Eval(program)
	if val.Type() == object.ERROR_OBJ {
		errorObj := val.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		for evaluator.ErrorTokens.Len() > 0 {
			buf.WriteString(lexer.GetErrorLineMessage(evaluator.ErrorTokens.PopBack()))
			buf.WriteByte('\n')
		}
		out.WriteString(fmt.Sprintf("EvaluatorError: %s", buf.String()))
		os.Exit(1)
	}
}`

	gomain := fmt.Sprintf("%s\n%s\n%s", header, entryPointPath, mainFunc)
	if isDebug {
		fmt.Println("gomain: -------------------------------------")
		fmt.Println(gomain)
	}

	renameOriginalMainGoFile := func() {
		err := os.Rename("main.go", "main.go.tmp")
		if err != nil {
			os.Stderr.WriteString("`main.go` rename failed to `main.go.tmp`. Error: " + err.Error() + "\n")
			os.Exit(1)
		}
	}
	writeMainGoFile := func(fdata string) {
		f, err := os.Create("main.go")
		if err != nil {
			os.Stderr.WriteString("failed to created `main.go` file. Error: " + err.Error() + "\n")
			os.Exit(1)
		}
		_, err = f.WriteString(fdata)
		if err != nil {
			os.Stderr.WriteString("failed to write file data to `main.go` file. Error: " + err.Error() + "\n")
			f.Close()
			os.Exit(1)
		}
		err = f.Close()
		if err != nil {
			os.Stderr.WriteString("failed to close `main.go` file. Error: " + err.Error() + "\n")
			os.Exit(1)
		}
	}
	// TODO: Return an error here instead of os.Exit so it can be handled (and we swap back main files)
	buildExe := func() {
		exeName := filepath.Base(fpath)
		cmd := []string{"go", "build", "-o", exeName + ".exe"}
		if runtime.GOOS == "windows" {
			winArgs := []string{"/c"}
			winArgs = append(winArgs, cmd...)
			output, err := exec.Command("cmd", winArgs...).CombinedOutput()
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf("failed to exec `%s`. Error: %s\n", strings.Join(winArgs, " "), err.Error()))
				os.Exit(1)
			}
			if len(output) == 0 {
				os.Stdout.WriteString("Successfully built `" + cmd[len(cmd)-1] + "` as Executable!\n")
			}
		} else {
			output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf("failed to exec `%s`. Error: %s\n", strings.Join(cmd, " "), err.Error()))
				os.Exit(1)
			}
			if len(output) == 0 {
				os.Stdout.WriteString("Successfully built `" + cmd[len(cmd)-1] + "` as Executable!\n")
			}
		}
	}
	removeMainGoFile := func() {
		err := os.Remove("main.go")
		if err != nil {
			os.Stderr.WriteString("failed to remove `main.go` file. Error: " + err.Error() + "\n")
			os.Exit(1)
		}
	}
	revertRenameOfOriginalGoFile := func() {
		err := os.Rename("main.go.tmp", "main.go")
		if err != nil {
			os.Stderr.WriteString("`main.go.tmp` rename failed to `main.go`. Error: " + err.Error())
			os.Exit(1)
		}
	}

	// These steps need to executed in this order
	renameOriginalMainGoFile()
	writeMainGoFile(gomain)
	buildExe()
	removeMainGoFile()
	revertRenameOfOriginalGoFile()
}
