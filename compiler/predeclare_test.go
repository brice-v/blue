package compiler

import (
	"strings"
	"testing"
)

func compileInput(input string) (*Compiler, error) {
	program := parse(input)
	comp := New()
	return comp, comp.Compile(program)
}

func TestPredeclareConsumesReservedSlot(t *testing.T) {
	symbolTable := NewSymbolTable()

	reserved := symbolTable.Predeclare("a", false)
	if reserved.Index != 0 {
		t.Fatalf("expected predeclaration to reserve slot 0, got %d", reserved.Index)
	}
	if reserved.Scope != GlobalScope {
		t.Fatalf("expected predeclaration in the root table to be global, got %s", reserved.Scope)
	}

	defined := symbolTable.Define("a", false)
	if defined.Index != reserved.Index {
		t.Fatalf("expected define to bind the reserved slot %d, got %d", reserved.Index, defined.Index)
	}
	if !defined.Equal(reserved) {
		t.Fatalf("expected define to produce %+v, got %+v", reserved, defined)
	}

	next := symbolTable.Define("b", false)
	if next.Index != 1 {
		t.Fatalf("expected next definition to get slot 1, got %d", next.Index)
	}
	if got := symbolTable.NumLocals(); got != 2 {
		t.Fatalf("expected 2 definitions, got %d", got)
	}
}

func TestPredeclareSkipsExistingDefinition(t *testing.T) {
	symbolTable := NewSymbolTable()
	existing := symbolTable.Define("a", false)
	symbolTable.Predeclare("a", false)
	if got := symbolTable.NumLocals(); got != 1 {
		t.Fatalf("predeclaring an already defined name should not allocate, got %d definitions", got)
	}
	resolved, ok := symbolTable.Resolve("a")
	if !ok {
		t.Fatal("expected to resolve 'a'")
	}
	if !resolved.Equal(existing) {
		t.Fatalf("expected %+v, got %+v", existing, resolved)
	}
}

func TestUseBeforeDefinitionInSameScopeFails(t *testing.T) {
	tests := []string{
		`x; var x = 5;`,
		`println(x); var x = 5;`,
		`foo(); fun foo() { 1 }`,
		`fun f() { y; var y = 2; }`,
		`fun f() { if true { z } var z = 3; }`,
	}
	for _, input := range tests {
		_, err := compileInput(input)
		if err == nil {
			t.Fatalf("expected compiler error for %q", input)
		}
		if !strings.Contains(err.Error(), "identifier not found") {
			t.Fatalf("expected 'identifier not found' for %q, got: %s", input, err.Error())
		}
	}
}

func TestUndefinedNameStillFails(t *testing.T) {
	_, err := compileInput(`nope;`)
	if err == nil {
		t.Fatal("expected compiler error")
	}
	if !strings.Contains(err.Error(), "identifier not found nope") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestDefinitionAfterUse(t *testing.T) {
	tests := []string{
		`fun main() { add(1, 2) } main(); fun add(a, b) { a + b }`,
		`fun double() { multiplier * 2 } var multiplier = 21 double()`,
		`fun run() { data.count() } var data = [1, 2, 3] run()`,
		`val x = 1 if true { val y = 2 x + y }`,
	}
	for _, input := range tests {
		_, err := compileInput(input)
		if err != nil {
			t.Fatalf("expected %q to compile, got: %s", input, err)
		}
	}
}

func TestClosureCannotReadLocalBeforeItsDeclaration(t *testing.T) {
	input := `
	fun outer() {
		var f = fun() { g }
		var g = 42
		f
	}
	`
	_, err := compileInput(input)
	if err == nil {
		t.Fatal("expected a closure to not read a local before it is declared")
	}
	if !strings.Contains(err.Error(), "identifier not found g") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestPredeclaredSlotsStayStableAcrossBlocks(t *testing.T) {
	input := `
	var total = 0
	for (var i = 0; i < 3; i += 1) {
		total += i
	}
	total

	var counter = 0
	for (item in [4, 5]) {
		counter += item
	}
	counter
	`
	comp, err := compileInput(input)
	if err != nil {
		t.Fatalf("expected to compile, got: %s", err)
	}
	if len(comp.Bytecode().Instructions) == 0 {
		t.Fatal("expected instructions")
	}
}
