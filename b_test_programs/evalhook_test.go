package b_program_test

import (
	"blue/compiler"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/vm"
	"os"
	"testing"
)

// Test-only glue: full-build binaries install vm.EvalHook from their own
// entrypoints (see cmd/util.go). Tests run against the raw VM so they
// install it here to exercise programs using eval/to_num/load.
func TestMain(m *testing.M) {
	vm.EvalHook = evalForTests
	os.Exit(m.Run())
}

func evalForTests(src string) object.Object {
	l := lexer.New(src, "<internal:string>")
	p := parser.New(l)
	prog := p.ParseProgram()
	if p.HasErrors() {
		return &object.Error{Message: "failed to `eval` string"}
	}
	c := compiler.New()
	if err := c.Compile(prog); err != nil {
		return &object.Error{Message: "compiler error in `eval` string: " + err.Error()}
	}
	vmInstance := vm.New(c.Bytecode())
	if err := vmInstance.Run(); err != nil {
		return &object.Error{Message: "vm error in `eval` string: " + err.Error()}
	}
	return vmInstance.LastPoppedStackElem()
}
