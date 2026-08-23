package vm

import (
	"blue/compiler"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"os"
	"testing"
)

// This file is the TEST-ONLY equivalent of the glue code full binaries use
// to install vm.EvalHook. Production entrypoints (cmd, wasmmain) install
// their own; minimal builds install none.

func TestMain(m *testing.M) {
	EvalHook = evalForTests
	os.Exit(m.Run())
}

// evalForTests mirrors what cmd installs for production builds.
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
	vm := New(c.Bytecode())
	if err := vm.Run(); err != nil {
		return &object.Error{Message: "vm error in `eval` string: " + err.Error()}
	}
	return vm.LastPoppedStackElem()
}
