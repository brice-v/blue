package b_program_test

import (
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fibEx = `fun fib(n) {
    if n < 2 {
        return n;
    }

    return fib(n-1) + fib(n-2);
}

fib(10);`

func TestAllProgramsInDirectory(t *testing.T) {
	files, err := os.ReadDir("./")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		executeBlueTestFile(f, t)
		// TODO: See if we can make our own execution environment for blue
		// that way the gos (global object store) can just be instantiated
		// for new test runs (in parallel)
		// ff := f
		// t.Run(ff.Name(), func(t *testing.T) {
		// 	t.Parallel()
		// 	executeBlueTestFile(ff, t)
		// })
	}
}

func executeBlueTestFile(f fs.DirEntry, t *testing.T) {
	if !strings.HasSuffix(f.Name(), ".b") {
		return
	}
	// Note: Comment out this defered func to see what the panic trace is
	defer func() {
		// recover from panic if one occured. Set err to nil otherwise.
		err := recover()
		if err != nil {
			t.Fatalf("PANIC in FILE `%s`: Error: %+v", f.Name(), err)
		}
	}()
	fpath, err := filepath.Abs(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	openFile, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer openFile.Close()

	data, err := io.ReadAll(openFile)
	if err != nil {
		t.Fatal(err)
	}
	stringData := string(data)
	if strings.HasPrefix(stringData, "# IGNORE") || strings.HasPrefix(stringData, "#IGNORE") {
		return
	}
	l := lexer.New(stringData, fpath)

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(os.Stderr, p.Errors())
		t.Fatalf("File `%s`: failed to parse", f.Name())
	}
	e := evaluator.New()
	obj := e.Eval(program)
	if obj == nil {
		t.Fatalf("File `%s`: evaluator returned nil", f.Name())
	}
	if obj.Type() == object.ERROR_OBJ {
		errorObj := obj.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		for e.ErrorTokens.Len() > 0 {
			buf.WriteString(lexer.GetErrorLineMessage(e.ErrorTokens.PopBack()))
			buf.WriteByte('\n')
		}
		t.Fatalf("File `%s`: evaluator returned error: %s", f.Name(), buf.String())
	}
	if obj.Inspect() != "true" {
		t.Fatalf("File `%s`: Did not return true as last statement. Failed", f.Name())
	}
}

func TestFibPerf(t *testing.T) {
	execString(t, fibEx)
}

func execString(t *testing.T, s string) {
	l := lexer.New(s, "<string>")

	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(os.Stderr, p.Errors())
		t.Fatalf("failed to parse string: %s", s)
	}
	e := evaluator.New()
	obj := e.Eval(program)
	if obj == nil {
		t.Fatalf("evaluator returned nil: %s", s)
	}
	if obj.Type() == object.ERROR_OBJ {
		errorObj := obj.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		for e.ErrorTokens.Len() > 0 {
			buf.WriteString(lexer.GetErrorLineMessage(e.ErrorTokens.PopBack()))
			buf.WriteByte('\n')
		}
		t.Fatalf("evaluator returned error: %s, %s", s, buf.String())
	}
}
