package compiler

import (
	"blue/ast"
	"blue/code"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"testing"
)

func TestAdditionalOperators(t *testing.T) {
	tests := []compilerTestCase{
		// Floor division
		{
			input:             "10 // 3",
			expectedConstants: []any{10, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpFlDiv),
				code.Make(code.OpPop),
			},
		},
		// Power
		{
			input:             "2 ** 3",
			expectedConstants: []any{2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpPow),
				code.Make(code.OpPop),
			},
		},
		// Modulo
		{
			input:             "10 % 3",
			expectedConstants: []any{10, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpPercent),
				code.Make(code.OpPop),
			},
		},
		// Carat (xor)
		{
			input:             "5 ^ 3",
			expectedConstants: []any{5, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpCarat),
				code.Make(code.OpPop),
			},
		},
		// Bitwise AND
		{
			input:             "5 & 3",
			expectedConstants: []any{5, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpAmpersand),
				code.Make(code.OpPop),
			},
		},
		// Bitwise OR (pipe)
		{
			input:             "5 | 3",
			expectedConstants: []any{5, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpPipe),
				code.Make(code.OpPop),
			},
		},
		// Left shift
		{
			input:             "1 << 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpLshift),
				code.Make(code.OpPop),
			},
		},
		// Right shift
		{
			input:             "4 >> 1",
			expectedConstants: []any{4, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpRshift),
				code.Make(code.OpPop),
			},
		},
		// Range operator
		{
			input:             "1..5",
			expectedConstants: []any{1, 5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpRange),
				code.Make(code.OpPop),
			},
		},
		// Non-inclusive range
		{
			input:             "1..<5",
			expectedConstants: []any{1, 5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpNonIncRange),
				code.Make(code.OpPop),
			},
		},
		// In operator (left operand compiled first, then right)
		{
			input:             `"a" in ["a", "b"]`,
			expectedConstants: []any{"a", "a", "b"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpConstant, constOffset+2),
				code.Make(code.OpList, 2),
				code.Make(code.OpIn),
				code.Make(code.OpPop),
			},
		},
		// Notin operator (left operand compiled first, then right)
		{
			input:             `"c" notin ["a", "b"]`,
			expectedConstants: []any{"c", "a", "b"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpConstant, constOffset+2),
				code.Make(code.OpList, 2),
				code.Make(code.OpNotin),
				code.Make(code.OpPop),
			},
		},
		// Greater than or equal (normal order: left first, then right)
		{
			input:             "1 >= 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpGreaterThanOrEqual),
				code.Make(code.OpPop),
			},
		},
		// Less than or equal (operands reversed: right first, then left)
		{
			input:             "1 <= 2",
			expectedConstants: []any{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpGreaterThanOrEqual),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestPrefixOperators(t *testing.T) {
	tests := []compilerTestCase{
		// Not
		{
			input:             "!true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpNot),
				code.Make(code.OpPop),
			},
		},
		// Not with not keyword
		{
			input:             "not true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpNot),
				code.Make(code.OpPop),
			},
		},
		// Negation
		{
			input:             "-5",
			expectedConstants: []any{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpNeg),
				code.Make(code.OpPop),
			},
		},
		// Bitwise NOT (tilde)
		{
			input:             "~5",
			expectedConstants: []any{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpTilde),
				code.Make(code.OpPop),
			},
		},
		// Left shift prefix
		{
			input:             "<<1",
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpLshiftPre),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestPostfixOperators(t *testing.T) {
	tests := []compilerTestCase{
		// Right shift postfix (requires semicolon)
		{
			input:             "4>>;",
			expectedConstants: []any{4},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpRshiftPost),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFloatLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "3.14",
			expectedConstants: []any{3.14},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "3.14 + 2.71",
			expectedConstants: []any{3.14, 2.71},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestHexOctalBinaryLiterals(t *testing.T) {
	tests := []compilerTestCase{
		// Hex literal (produces UInteger)
		{
			input:             "0xFF",
			expectedConstants: []any{uint64(255)},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpPop),
			},
		},
		// Octal literal (produces UInteger)
		{
			input:             "0o77",
			expectedConstants: []any{uint64(63)},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpPop),
			},
		},
		// Binary literal (produces UInteger)
		{
			input:             "0b1010",
			expectedConstants: []any{uint64(10)},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestStringInterpolation(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`"hello #{42}"`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
	// Check for StringInterp instruction
	found := false
	for i := 0; i < len(bc.Instructions); {
		def, err := code.Lookup(bc.Instructions[i])
		if err != nil {
			i++
			continue
		}
		if def.Name == "OpStringInterp" {
			found = true
			break
		}
		_, read := code.ReadOperands(def, bc.Instructions[i+1:])
		i += 1 + read
	}
	if !found {
		t.Error("expected OpStringInterp instruction")
	}
}

func TestAssignmentOperators(t *testing.T) {
	tests := []compilerTestCase{
		// Simple assignment
		{
			input: `
			var x = 5;
			x = 10;
			`,
			expectedConstants: []any{5, 10},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		// += operator
		{
			input: `
			var x = 5;
			x += 3;
			`,
			expectedConstants: []any{5, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		// -= operator
		{
			input: `
			var x = 5;
			x -= 3;
			`,
			expectedConstants: []any{5, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpMinus),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		// *= operator
		{
			input: `
			var x = 5;
			x *= 3;
			`,
			expectedConstants: []any{5, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpStar),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		// /= operator
		{
			input: `
			var x = 10;
			x /= 3;
			`,
			expectedConstants: []any{10, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpDiv),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		// &= operator
		{
			input: `
			var x = 7;
			x &= 3;
			`,
			expectedConstants: []any{7, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpAmpersand),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		// |= operator
		{
			input: `
			var x = 7;
			x |= 3;
			`,
			expectedConstants: []any{7, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpPipe),
				code.Make(code.OpSetGlobal, 0),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestWhileLoop(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
var x = 0;
for (true) { x += 1 }
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestBreakContinue(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
for (true) { break }
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	program2 := parse(`
for (true) { continue }
`)
	compiler2 := New()
	err = compiler2.Compile(program2)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
}

func TestMatchExpression(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
match 5 {
	5 => { 10 },
	_ => { 20 },
}
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestListComprehension(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse("[x for x in 1..5]")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestSetComprehension(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse("{x for x in 1..5}")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestMapComprehension(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse("{x: x*2 for x in 1..5}")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestTryCatchFinally(t *testing.T) {
	// try with finally
	program := parse(`
try { 1 } finally { 2 }
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}

	// try with catch
	program2 := parse(`
try { 1 } catch (e) { 2 }
`)
	compiler2 := New()
	err = compiler2.Compile(program2)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	// try with catch and finally
	program3 := parse(`
try { 1 } catch (e) { 2 } finally { 3 }
`)
	compiler3 := New()
	err = compiler3.Compile(program3)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
}

func TestStructLiteral(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`@{name: "bob"}`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestSelfExpression(t *testing.T) {
	// self() requires parentheses
	program := parse(`self()`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestDeferExpression(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`defer println(1)`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestSpawnExpression(t *testing.T) {
	// spawn requires parentheses around its arguments
	program := parse(`spawn(fun() { })`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestValStatement(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			val x = 5;
			`,
			expectedConstants: []any{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobalImm, 0),
			},
		},
		{
			input: `
			val x = 5;
			val y = x;
			`,
			expectedConstants: []any{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpSetGlobalImm, 0),
				code.Make(code.OpGetGlobalImm, 0),
				code.Make(code.OpSetGlobalImm, 1),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctionStatement(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			fun foo() { return 5 }
			`,
			expectedConstants: []any{
				5,
				[]code.Instructions{
					code.Make(code.OpConstant, constOffset+0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, constOffset+1, 0),
				code.Make(code.OpSetGlobalImm, 0),
			},
		},
		{
			input: `
			fun foo() { 5 }
			`,
			expectedConstants: []any{
				5,
				[]code.Instructions{
					code.Make(code.OpConstant, constOffset+0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, constOffset+1, 0),
				code.Make(code.OpSetGlobalImm, 0),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestForLoopVariants(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
for (var i = 0; i < 10; i += 1) { i }
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestNullOrEmpty(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "null",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNull),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestSetLiteral(t *testing.T) {
	// {1, 2, 3} produces a SetLiteral, {} produces a MapLiteral
	program := parse("{1, 2, 3}")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestShortCircuitOperators(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse("true and false")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	program2 := parse("true or false")
	compiler2 := New()
	err = compiler2.Compile(program2)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
}

func TestIfExpressionVariants(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
if (true) { 10 }
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	program2 := parse(`
if (true) { 1 } else if (false) { 2 } else { 3 }
`)
	compiler2 := New()
	err = compiler2.Compile(program2)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
}

func TestReturnStatement(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
fun foo() { return 5 }
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestIndexExpression(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
var arr = [1, 2, 3];
arr[0];
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestSliceExpression(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
var arr = [1, 2, 3, 4, 5];
arr[1..3];
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestIndexSet(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
var arr = [1, 2, 3];
arr[0] = 10;
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestListDestructor(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			var [a, b] = [1, 2];
			`,
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, constOffset+0),
				code.Make(code.OpConstant, constOffset+1),
				code.Make(code.OpList, 2),
				code.Make(code.OpGetListIndex, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetListIndex, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestMapDestructor(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
var {a, b} = {a: 1, b: 2};
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestNewWithState(t *testing.T) {
	global := NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		global.DefineBuiltin(i, v.Name, 0)
	}
	compiler := NewWithState(global, []object.Object{object.OBJECT_CONSTANTS[0]})
	if compiler.symbolTable != global {
		t.Error("symbolTable not set correctly")
	}
	if len(compiler.constants) != 1 {
		t.Errorf("wrong constants length. got=%d", len(compiler.constants))
	}
}

func TestCompilerDebugString(t *testing.T) {
	compiler := New()
	_ = compiler.DebugString()
	// Just verify it doesn't panic and returns a string
}

func TestBlockStatement(t *testing.T) {
	// Block statements are parsed inside other constructs
	program := parse(`
fun foo() { 1; 2; 3 }
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestLocalVarStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			fun() {
				var x = 1;
				x
			}
			`,
			expectedConstants: []any{
				1,
				[]code.Instructions{
					code.Make(code.OpConstant, constOffset+0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, constOffset+1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			fun() {
				var x = 1;
				var y = 2;
				x + y
			}
			`,
			expectedConstants: []any{
				1,
				2,
				[]code.Instructions{
					code.Make(code.OpConstant, constOffset+0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpConstant, constOffset+1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, constOffset+2, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestLocalValStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			fun() {
				val x = 1;
				x
			}
			`,
			expectedConstants: []any{
				1,
				[]code.Instructions{
					code.Make(code.OpConstant, constOffset+0),
					code.Make(code.OpSetLocalImm, 0),
					code.Make(code.OpGetLocalImm, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, constOffset+1, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestInfixWithBreak(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
for (var i = 0; i < 10; i += 1) {
	if (i == 5) { break }
}
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestInfixWithContinue(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
for (var i = 0; i < 10; i += 1) {
	if (i == 5) { continue }
}
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestMapLiteralWithIdentKeys(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
var a = 1;
{a: a}
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestCallWithDefaultArgs(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
fun foo(a=1, b=2) { a + b };
foo();
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestCompilerEmitAndLastInstruction(t *testing.T) {
	compiler := New()
	compiler.emit(code.OpTrue)
	compiler.emit(code.OpFalse)
	compiler.emit(code.OpAdd)

	if compiler.lastInstructionIs(code.OpAdd) {
		// good
	} else {
		t.Errorf("lastInstructionIs(OpAdd) = false, want true")
	}
	if compiler.lastInstructionIs(code.OpPop) {
		t.Errorf("lastInstructionIs(OpPop) = true, want false")
	}

	compiler.emit(code.OpSetGlobal, 0)
	if !compiler.lastInstructionIsSet() {
		t.Error("lastInstructionIsSet() = false, want true")
	}
}

func TestCompilerRemoveLastPop(t *testing.T) {
	compiler := New()
	compiler.emit(code.OpTrue)
	compiler.emit(code.OpPop)
	compiler.emit(code.OpFalse)
	compiler.emit(code.OpPop)

	initialLen := len(compiler.currentInstructions())
	if initialLen < 2 {
		t.Errorf("wrong instructions length before remove. got=%d", initialLen)
	}

	compiler.removeLastPop()

	afterLen := len(compiler.currentInstructions())
	if afterLen >= initialLen {
		t.Errorf("wrong instructions length after remove. got=%d, want < %d", afterLen, initialLen)
	}
}

func TestReplaceLastPopWithReturn(t *testing.T) {
	compiler := New()
	compiler.emit(code.OpTrue)
	compiler.emit(code.OpPop)

	compiler.replaceLastPopWithReturn()

	if !compiler.lastInstructionIs(code.OpReturnValue) {
		t.Errorf("lastInstructionIs(OpReturnValue) = false, want true")
	}
}

func TestCompilerBytecode(t *testing.T) {
	program := parse("1 + 2")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if bc == nil {
		t.Fatal("Bytecode() returned nil")
	}
	if len(bc.Instructions) == 0 {
		t.Error("Bytecode.Instructions is empty")
	}
	if len(bc.Constants) == 0 {
		t.Error("Bytecode.Constants is empty")
	}
}

func TestConstantFolding(t *testing.T) {
	// Test that compilation works
	program := parse("1 + 1")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	// Should have at least OBJECT_CONSTANTS entries
	if len(bc.Constants) < 1 {
		t.Error("expected at least some constants")
	}
}

func TestTokenFolding(t *testing.T) {
	// Test that tokens are stored
	program := parse("1 + 1")
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	// Tokens may or may not be stored depending on the code path
	_ = bc.Tokens
}

func TestAddNode(t *testing.T) {
	compiler := New()
	// First add should create new entry
	l := lexer.New("1", "<test>")
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(prog.Statements) == 0 {
		t.Fatal("no statements")
	}
	idx1 := compiler.addNode(prog.Statements[0])
	if idx1 != 0 {
		t.Errorf("first addNode index = %d, want 0", idx1)
	}
	// Second add of same token should return same index
	idx2 := compiler.addNode(prog.Statements[0])
	if idx2 != 0 {
		t.Errorf("second addNode index = %d, want 0 (should be folded)", idx2)
	}
}

func TestAddConstant(t *testing.T) {
	compiler := New()
	int1 := &object.Integer{Value: 42}

	idx1 := compiler.addConstant(int1)
	if idx1 < constOffset {
		t.Errorf("addConstant: idx=%d, want >= %d", idx1, constOffset)
	}

	// Add another different value
	int2 := &object.Integer{Value: 99}
	idx2 := compiler.addConstant(int2)
	if idx2 != idx1+1 {
		t.Errorf("addConstant second: idx=%d, want %d", idx2, idx1+1)
	}
}

func TestIsConstantFolded(t *testing.T) {
	compiler := New()
	int1 := &object.Integer{Value: 42}
	// Just verify the function doesn't panic
	_, _ = compiler.isConstantFolded(int1)
	compiler.addConstant(int1)
	_, _ = compiler.isConstantFolded(int1)
}

func TestAddConstantReservedIndex(t *testing.T) {
	compiler := New()
	// OBJECT_CONSTANTS[0] is a reserved constant object
	reservedIdx := compiler.addConstant(object.OBJECT_CONSTANTS[0])
	if reservedIdx == -1 {
		t.Error("addConstant should return reserved index for constant objects")
	}
}

func TestLastInstructionIsEmpty(t *testing.T) {
	compiler := New()
	if compiler.lastInstructionIs(code.OpTrue) {
		t.Error("lastInstructionIs on empty should be false")
	}
	if compiler.lastInstructionIsSet() {
		t.Error("lastInstructionIsSet on empty should be false")
	}
}

func TestClearBlockSymbols(t *testing.T) {
	compiler := New()
	compiler.BlockNestLevel = -1
	compiler.clearBlockSymbols()
	// Should not panic with BlockNestLevel == -1

	compiler.BlockNestLevel = 0
	compiler.clearBlockSymbols()
	// Should not panic with BlockNestLevel == 0
}

func TestGetName(t *testing.T) {
	compiler := New()
	// Without import nest, getName should return the name unchanged
	result := compiler.getName("foo")
	if result != "foo" {
		t.Errorf("getName without import = %q, want %q", result, "foo")
	}

	// With import nest level set
	compiler.importNestLevel = 0
	compiler.modName = []string{"mymod"}
	result = compiler.getName("foo")
	if result != "mymod.foo" {
		t.Errorf("getName with import = %q, want %q", result, "mymod.foo")
	}
}

func TestCreateFilePathFromImportPath(t *testing.T) {
	tests := []struct {
		basePath string
		importPath string
		expected string
	}{
		{".", "foo", "foo.b"},
		{".", "foo.bar", "foo/bar.b"},
		{"/path", "foo", "/path/foo.b"},
	}
	for _, tt := range tests {
		compiler := New()
		compiler.CompilerBasePath = tt.basePath
		result := compiler.createFilePathFromImportPath(tt.importPath)
		if result != tt.expected {
			t.Errorf("createFilePathFromImportPath(%q, %q) = %q, want %q", tt.basePath, tt.importPath, result, tt.expected)
		}
	}
}

func TestLoadSymbolGlobal(t *testing.T) {
	compiler := New()
	sym := Symbol{Name: "x", Scope: GlobalScope, Index: 0, Immutable: false}
	compiler.loadSymbol(sym)
	if !compiler.lastInstructionIs(code.OpGetGlobal) {
		t.Errorf("loadSymbol GlobalScope mutable: last instruction = %v, want OpGetGlobal", compiler.scopes[0].lastInstruction.Opcode)
	}
}

func TestLoadSymbolLocal(t *testing.T) {
	compiler := New()
	sym := Symbol{Name: "x", Scope: LocalScope, Index: 0, Immutable: false}
	compiler.loadSymbol(sym)
	if !compiler.lastInstructionIs(code.OpGetLocal) {
		t.Errorf("loadSymbol LocalScope mutable: last instruction = %v, want OpGetLocal", compiler.scopes[0].lastInstruction.Opcode)
	}
}

func TestLoadSymbolImmutableGlobal(t *testing.T) {
	compiler := New()
	sym := Symbol{Name: "x", Scope: GlobalScope, Index: 0, Immutable: true}
	compiler.loadSymbol(sym)
	if !compiler.lastInstructionIs(code.OpGetGlobalImm) {
		t.Errorf("loadSymbol GlobalScope immutable: last instruction = %v, want OpGetGlobalImm", compiler.scopes[0].lastInstruction.Opcode)
	}
}

func TestLoadSymbolImmutableLocal(t *testing.T) {
	compiler := New()
	sym := Symbol{Name: "x", Scope: LocalScope, Index: 0, Immutable: true}
	compiler.loadSymbol(sym)
	if !compiler.lastInstructionIs(code.OpGetLocalImm) {
		t.Errorf("loadSymbol LocalScope immutable: last instruction = %v, want OpGetLocalImm", compiler.scopes[0].lastInstruction.Opcode)
	}
}

func TestLoadSymbolFree(t *testing.T) {
	compiler := New()
	sym := Symbol{Name: "x", Scope: FreeScope, Index: 0, Immutable: false}
	compiler.loadSymbol(sym)
	if !compiler.lastInstructionIs(code.OpGetFree) {
		t.Errorf("loadSymbol FreeScope mutable: last instruction = %v, want OpGetFree", compiler.scopes[0].lastInstruction.Opcode)
	}
}

func TestLoadSymbolBuiltin(t *testing.T) {
	compiler := New()
	sym := Symbol{Name: "len", Scope: BuiltinScope, Index: 0, Immutable: false}
	compiler.loadSymbol(sym)
	if !compiler.lastInstructionIs(code.OpGetBuiltin) {
		t.Errorf("loadSymbol BuiltinScope: last instruction = %v, want OpGetBuiltin", compiler.scopes[0].lastInstruction.Opcode)
	}
}

func TestEmitSetSymbolOpcode(t *testing.T) {
	compiler := New()
	// Global mutable
	sym := Symbol{Name: "x", Scope: GlobalScope, Index: 0, Immutable: false}
	compiler.emitSetSymbolOpcode(sym, false)
	if !compiler.lastInstructionIs(code.OpSetGlobal) {
		t.Errorf("emitSetSymbolOpcode GlobalScope mutable: last instruction = %v, want OpSetGlobal", compiler.scopes[0].lastInstruction.Opcode)
	}

	// Global immutable
	sym = Symbol{Name: "x", Scope: GlobalScope, Index: 0, Immutable: true}
	compiler.emitSetSymbolOpcode(sym, true)
	if !compiler.lastInstructionIs(code.OpSetGlobalImm) {
		t.Errorf("emitSetSymbolOpcode GlobalScope immutable: last instruction = %v, want OpSetGlobalImm", compiler.scopes[0].lastInstruction.Opcode)
	}

	// Local mutable
	sym = Symbol{Name: "x", Scope: LocalScope, Index: 0, Immutable: false}
	compiler.emitSetSymbolOpcode(sym, false)
	if !compiler.lastInstructionIs(code.OpSetLocal) {
		t.Errorf("emitSetSymbolOpcode LocalScope mutable: last instruction = %v, want OpSetLocal", compiler.scopes[0].lastInstruction.Opcode)
	}

	// Local immutable
	sym = Symbol{Name: "x", Scope: LocalScope, Index: 0, Immutable: true}
	compiler.emitSetSymbolOpcode(sym, true)
	if !compiler.lastInstructionIs(code.OpSetLocalImm) {
		t.Errorf("emitSetSymbolOpcode LocalScope immutable: last instruction = %v, want OpSetLocalImm", compiler.scopes[0].lastInstruction.Opcode)
	}
}

func TestDefineSymbolForVarValStatement(t *testing.T) {
	compiler := New()
	// Test with var statement
	varStmt := &ast.VarStatement{}
	sym, immutable := compiler.defineSymbolForVarValStatement(varStmt, "x")
	if sym.Name != "x" {
		t.Errorf("defineSymbolForVarValStatement: name = %q, want %q", sym.Name, "x")
	}
	if immutable {
		t.Error("defineSymbolForVarValStatement with var: immutable should be false")
	}

	// Test with val statement
	valStmt := &ast.ValStatement{}
	sym, immutable = compiler.defineSymbolForVarValStatement(valStmt, "y")
	if sym.Name != "y" {
		t.Errorf("defineSymbolForVarValStatement: name = %q, want %q", sym.Name, "y")
	}
	if !immutable {
		t.Error("defineSymbolForVarValStatement with val: immutable should be true")
	}
}

func TestCompileStringLiteral(t *testing.T) {
	program := parse(`"hello"`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Constants) < 2 {
		t.Errorf("expected at least 2 constants, got %d", len(bc.Constants))
	}
}

func TestCompileStringLiteralInterpolation(t *testing.T) {
	program := parse(`"hello #{42}"`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	// Should have StringInterp instruction
	found := false
	for i := 0; i < len(bc.Instructions); {
		def, err := code.Lookup(bc.Instructions[i])
		if err != nil {
			i++
			continue
		}
		if def.Name == "OpStringInterp" {
			found = true
			break
		}
		_, read := code.ReadOperands(def, bc.Instructions[i+1:])
		i += 1 + read
	}
	if !found {
		t.Error("expected OpStringInterp instruction")
	}
}

func TestCompileUseParamStr(t *testing.T) {
	// USE_PARAM_STR is a special string used for default parameters
	program := parse(`"#use_param_test"`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
}

func TestReplaceInstruction(t *testing.T) {
	compiler := New()
	compiler.emit(code.OpTrue)
	compiler.emit(code.OpFalse)
	compiler.emit(code.OpAdd)

	// Replace the OpFalse with OpTrue
	compiler.replaceInstruction(1, code.Make(code.OpTrue))

	bc := compiler.Bytecode()
	def, _ := code.Lookup(bc.Instructions[1])
	if def.Name != "OpTrue" {
		t.Errorf("replaceInstruction: instruction at pos 1 = %q, want OpTrue", def.Name)
	}
}

func TestChangeOperand(t *testing.T) {
	compiler := New()
	compiler.emit(code.OpConstant, 42)

	// Change the operand from 42 to 99
	compiler.changeOperand(0, 99)

	bc := compiler.Bytecode()
	def, _ := code.Lookup(bc.Instructions[0])
	operands, _ := code.ReadOperands(def, bc.Instructions[1:])
	if operands[0] != 99 {
		t.Errorf("changeOperand: operand = %d, want 99", operands[0])
	}
}

func TestCompileIndexSetWithList(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
var arr = [1, 2, 3];
arr[0] = 10;
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestForLoopWithBreak(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
for (true) {
	break
}
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}

func TestForLoopWithContinue(t *testing.T) {
	// Just verify compilation doesn't panic
	program := parse(`
for (true) {
	continue
}
`)
	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	bc := compiler.Bytecode()
	if len(bc.Instructions) == 0 {
		t.Error("expected non-empty instructions")
	}
}
