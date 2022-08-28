package b_program_test

import (
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllProgramsInDirectory(t *testing.T) {
	files, err := os.ReadDir("./")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".b") {
			continue
		}
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
		l := lexer.New(string(data))

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
			t.Fatalf("File `%s`: evaluator returned error: %s", f.Name(), obj)
		}
		if obj.Inspect() != "true" {
			t.Fatalf("File `%s`: Did not return true as last statement. Failed", f.Name())
		}
	}
}
