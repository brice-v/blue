package cmd

import (
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"blue/token"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// fileExists is a helper function to check if the fpath given
// exists and if not return false and the error
func fileExists(fpath string) (bool, error) {
	_, err := os.Stat(fpath)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, errors.New("filepath does not exist")
	}
	return false, err
}

// isValidFile checks if the second argument is a valid file
func isValidFile() bool {
	if !(len(os.Args) >= 2) {
		os.Stderr.WriteString("Filepath not given")
		return false
	}
	ok, err := fileExists(os.Args[2])
	if !ok {
		msg := fmt.Sprintf("Unexpected error when trying to open %s | Error: %s | Exiting...\n", os.Args[1], err)
		os.Stderr.WriteString(msg)
		return false
	}
	return true
}

func isValidFpath(fpath string) bool {
	ok, err := fileExists(fpath)
	if !ok {
		msg := fmt.Sprintf("Unexpected error when trying to open %s | Error: %s | Exiting...\n", os.Args[1], err)
		os.Stderr.WriteString(msg)
		return false
	}
	return true
}

// isValidFileForEval checks if the first argument is a valid file
func isValidFileForEval() bool {
	if !(len(os.Args) >= 2) {
		os.Stderr.WriteString("Filepath not given")
		return false
	}
	ok, err := fileExists(os.Args[1])
	if !ok {
		msg := fmt.Sprintf("Unexpected error when trying to open %s | Error: %s | Exiting...\n", os.Args[1], err)
		os.Stderr.WriteString(msg)
		return false
	}
	return true
}

// lexCurrentFile lex's the second argument as a file
func lexCurrentFile() {
	fpath := os.Args[2]
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("Error trying to readfile `%s` | Error: %s", fpath, err)
	}

	l := lexer.New(string(data), fpath)

	for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
		fmt.Printf("%+v\n", tok)
	}
}

// parseCurrentFile parse's the second argument as a file
func parseCurrentFile() {
	fpath, out := os.Args[2], os.Stdout
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("Error trying to readfile `%s` | Error: %s", fpath, err)
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

// evalCurrentFile parse's the second argument as a file
func evalCurrentFile() {
	fpath, out := os.Args[2], os.Stdout
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("Error trying to readfile `%s` | Error: %s", fpath, err)
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
		out.WriteString(fmt.Sprintf("EvaluatorError: %s", buf.String()))
		os.Exit(1)
	}
	// NOTE: This could be used for debugging programs return values
	// if evaluated != nil {
	// 	os.Stdout.WriteString(evaluated.Inspect() + "\n")
	// }
}

// evalFile parse's the second argument as a file
func evalFile() {
	fpath, out := os.Args[1], os.Stdout
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("Error trying to readfile `%s` | Error: %s", fpath, err)
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
		out.WriteString(fmt.Sprintf("EvaluatorError: %s", buf.String()))
		os.Exit(1)
	}
	// NOTE: This could be used for debugging programs return values
	// if evaluated != nil {
	// 	os.Stdout.WriteString(evaluated.Inspect() + "\n")
	// }
}

// bundleCurrentFile parse's the second argument as a file
// and bundles the interpreter with the code into an executable
func bundleCurrentFile(fpath string, isDebug bool) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("Error trying to readfile `%s` | Error: %s", fpath, err)
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
