package cmd

import (
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"blue/token"
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
		return false, errors.New("Filepath does not exist")
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

	l := lexer.New(string(data))

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

	l := lexer.New(string(data))

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

	l := lexer.New(string(data))

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
		err := val.(*object.Error).Message
		out.WriteString("EvaluatorError: " + err + "\n")
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

	l := lexer.New(string(data))

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
		err := val.(*object.Error).Message
		out.WriteString("EvaluatorError: " + err + "\n")
		os.Exit(1)
	}
	// NOTE: This could be used for debugging programs return values
	// if evaluated != nil {
	// 	os.Stdout.WriteString(evaluated.Inspect() + "\n")
	// }
}

// bundleCurrentFile parse's the second argument as a file
// and bundles the interpreter with the code into an executable
func bundleCurrentFile() {
	fpath := os.Args[2]
	data, err := os.ReadFile(fpath)
	if err != nil {
		log.Fatalf("Error trying to readfile `%s` | Error: %s", fpath, err)
	}

	d := string(data)
	fmt.Println("File Name: '" + fpath + "', Data: ")
	fmt.Printf("`%s`\n\n", d)

	header := `package main

import (
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"os"
	"path/filepath"
	_ "embed"
)

var out = os.Stderr
`
	input := fmt.Sprintf("//go:embed %s\nvar input string\n", fpath)
	mainFunc := `func main() {
	l := lexer.New(input)
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
		err := val.(*object.Error).Message
		out.WriteString("EvaluatorError: " + err + "\n")
		os.Exit(1)
	}
}`

	gomain := fmt.Sprintf("%s\n%s\n%s", header, input, mainFunc)
	fmt.Println("gomain: -------------------------------------")
	fmt.Println(gomain)

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
				os.Stderr.WriteString("got 0 bytes from `" + strings.Join(cmd, " ") + "` output, not sure if thats exepected... continuing...\n")
			}
		} else {
			output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf("failed to exec `%s`. Error: %s\n", strings.Join(cmd, " "), err.Error()))
				os.Exit(1)
			}
			if len(output) == 0 {
				os.Stderr.WriteString("got 0 bytes from `" + strings.Join(cmd, " ") + "` output, not sure if thats exepected... continuing...\n")
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
