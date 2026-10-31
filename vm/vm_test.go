package vm

import (
	"blue/bluec"
	"blue/compiler"
	"blue/object"
	"fmt"
	"strings"
	"testing"

	"blue/ast"
	"blue/lexer"
	"blue/parser"
)

func parse(input string) *ast.Program {
	l := lexer.New(input, "<test>")
	p := parser.New(l)
	return p.ParseProgram()
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}
	return nil
}

type vmTestCase struct {
	input    string
	expected any
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()
	for i, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error in test %d: %s", i+1, err)
		}
		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm  in test %d: error: %s", i+1, err)
		}
		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected any, actual object.Object) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case []int:
		array, ok := actual.(*object.List)
		if !ok {
			t.Errorf("object not List: %T (%+v)", actual, actual)
			return
		}
		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(array.Elements))
			return
		}
		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case object.OrderedMap2[object.HashKey, object.MapPair]:
		m, ok := actual.(*object.Map)
		if !ok {
			t.Errorf("object is not map. got=%T (%+v)", actual, actual)
			return
		}
		if m.Pairs.Len() != expected.Len() {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d", expected.Len(), m.Pairs.Len())
			return
		}
		for _, k := range expected.Keys {
			expectedKey := k
			expectedValue, _ := expected.Get(k)
			pair, ok := m.Pairs.Get(expectedKey)
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}
			err := testIntegerObject(expectedValue.Value.(*object.Integer).Value, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case *object.Error:
		errObj, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", actual, actual)
			return
		}
		if errObj.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, errObj.Message)
		}
	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 * (2 + 10)", 60},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	runVmTests(t, tests)
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
	}
	return nil
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"!true", false},
		{"not false", true},
		{"not 5", false},
		{"not not true", true},
		{"not not false", false},
		{"not not 5", true},
		{"!(if (false) { 5; })", true},
	}
	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 } ", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (not true) { 1 } else if (3>5) { 2 } else if (8>4) { 3 } else { 4 }", 3},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}
	runVmTests(t, tests)
}

func TestGlobalVarStatements(t *testing.T) {
	tests := []vmTestCase{
		{"var one = 1; one", 1},
		{"var one = 1; var two = 2; one + two", 3},
		{"var one = 1; var two = one + one; one + two", 3},
		{"val one = 1; one", 1},
		{"val one = 1; val two = 2; one + two", 3},
		{"val one = 1; val two = one + one; one + two", 3},
	}
	runVmTests(t, tests)
}

func TestListLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}
	runVmTests(t, tests)
}

func TestMapLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			"{}", object.NewPairsMap(),
		},
		{
			"{1: 2, 2: 3}",
			createMapPairsWithKeysAndValues([]int64{1, 2}, []int64{2, 3}),
		},
	}
	runVmTests(t, tests)
}

func createMapPairsWithKeysAndValues(keys, values []int64) object.OrderedMap2[object.HashKey, object.MapPair] {
	if len(keys) != len(values) {
		panic("bad testcase")
	}
	pairs := object.NewPairsMapWithSize(len(keys))
	for i := range len(keys) {
		keyObj := &object.Integer{Value: keys[i]}
		hk := object.HashKey{Type: object.INTEGER_OBJ, Value: object.HashObject(keyObj)}
		mp := object.MapPair{Key: keyObj, Value: &object.Integer{Value: values[i]}}
		pairs.Set(hk, mp)
	}
	return pairs
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", object.NULL},
		{"[1, 2, 3][99]", object.NULL},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", object.NULL},
		{"{}[0]", object.NULL},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			var fivePlusTen = fun() { 5 + 10; };
			fivePlusTen();
			`,
			expected: 15,
		},
		{
			input: `
			var one = fun() { 1; };
			var two = fun() { 2; };
			one() + two()
			`,
			expected: 3,
		},
		{
			input: `
			var a = fun() { 1 };
			var b = fun() { a() + 1 };
			var c = fun() { b() + 1 };
			c();
			`,
			expected: 3,
		},
	}
	runVmTests(t, tests)
}

func TestFunctionsWithReturnStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			var earlyExit = fun() { return 99; 100; };
			earlyExit();
			`,
			expected: 99,
		},
		{
			input: `
			var earlyExit = fun() { return 99; return 100; };
			earlyExit();
			`,
			expected: 99,
		},
	}
	runVmTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			var noReturn = fun() { };
			noReturn();
			`,
			expected: object.NULL,
		},
		{
			input: `
			var noReturn = fun() { };
			var noReturnTwo = fun() { noReturn(); };
			noReturn();
			noReturnTwo();
			`,
			expected: object.NULL,
		},
	}
	runVmTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			var returnsOne = fun() { 1; };
			var returnsOneReturner = fun() { returnsOne; };
			returnsOneReturner()();
			`,
			expected: 1,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			var one = fun() { var one = 1; one };
			one();
			`,
			expected: 1,
		},
		{
			input: `
			var oneAndTwo = fun() { var one = 1; var two = 2; one + two; };
			oneAndTwo();
			`,
			expected: 3,
		}, {
			input: `
			var oneAndTwo = fun() { var one = 1; var two = 2; one + two; };
			var threeAndFour = fun() { var three = 3; var four = 4; three + four; };
			oneAndTwo() + threeAndFour();
			`,
			expected: 10,
		},
		{
			input: `
			var firstFoobar = fun() { var foobar = 50; foobar; };
			var secondFoobar = fun() { var foobar = 100; foobar; };
			firstFoobar() + secondFoobar();
			`,
			expected: 150,
		},
		{
			input: `
			var globalSeed = 50;
			var minusOne = fun() {
				var num = 1;
				globalSeed - num;
			}
			var minusTwo = fun() {
				var num = 2;
				globalSeed - num;
			}
			minusOne() + minusTwo();
			`,
			expected: 97,
		},
		{
			input: `fun fib(n) {
				if n < 2 {
					return n;
				}

				return fib(n-1) + fib(n-2);
			}
			fib(28);`,
			expected: 317811,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			var identity = fun(a) { a; };
			identity(4);
			`,
			expected: 4,
		},
		{
			input: `
			var sum = fun(a, b) { a + b; };
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			var sum = fun(a, b) {
				var c = a + b;
				c;
			};
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			var sum = fun(a, b) {
				var c = a + b;
				c;
			};
			sum(1, 2) + sum(3, 4);`,
			expected: 10,
		},
		{
			input: `
			var sum = fun(a, b) {
				var c = a + b;
				c;
			};
			var outer = fun() {
				sum(1, 2) + sum(3, 4);
			};
			outer();
			`,
			expected: 10,
		},
	}
	runVmTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `fun() { 1; }(1);`,
			expected: `wrong number of arguments: want=0, got=1`,
		},
		{
			input:    `fun(a) { a; }();`,
			expected: `wrong number of arguments: want=1, got=0`,
		},
		{
			input:    `fun(a, b) { a + b; }(1);`,
			expected: `wrong number of arguments: want=2, got=1`,
		},
	}
	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		vm := New(comp.Bytecode())
		err = vm.Run()
		if err == nil {
			t.Fatalf("expected VM error but resulted in none.")
		}
		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error: want=%q, got=%q", tt.expected, err)
		}
	}
}

func TestCallingFunctionsWithWrongArgumentsRecordsErrorTrace(t *testing.T) {
	input := `fun f(a) {
		a;
	}
	f();`
	program := parse(input)
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	v := New(comp.Bytecode())
	err := v.Run()
	if err == nil {
		t.Fatal("expected VM error but resulted in none.")
	}
	if !strings.Contains(err.Error(), "wrong number of arguments: want=1, got=0") {
		t.Fatalf("unexpected VM error: %q", err.Error())
	}
	if len(v.TokensForErrorTrace) == 0 {
		t.Fatal("expected error trace tokens to be recorded, got none")
	}
}

func TestCallingFunctionsWithWrongArgumentsInsideTryIsCaught(t *testing.T) {
	input := `fun f(a) {
		a;
	}
	try {
		f();
	} catch (e) {
		assert(e == "wrong number of arguments: want=1, got=0");
	};`
	program := parse(input)
	comp := compiler.NewFromCore()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	v := New(comp.Bytecode())
	if err := v.Run(); err != nil {
		t.Fatalf("expected error to be caught, got: %s", err.Error())
	}
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			var newClosure = fun(a) {
				fun() { a; };
			};
			var closure = newClosure(99);
			closure();
			`,
			expected: 99,
		},
	}
	runVmTests(t, tests)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`print("hello", "world!")`, object.NULL},
		{`push([], 1)`, 1},
	}
	runVmTests(t, tests)
}

func TestStringOperations(t *testing.T) {
	tests := []vmTestCase{
		{`"hello" + " " + "world"`, ""},
		{`"abc" == "abc"`, true},
		{`"abc" == "xyz"`, false},
		{`"abc" != "xyz"`, true},
		{`"abc" != "abc"`, false},
	}
	runVmTests(t, tests)
}

func TestFloatArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{`1.5 + 2.5`, 4.0},
		{`10.0 - 3.5`, 6.5},
		{`2.5 * 4.0`, 10.0},
		{`15.0 / 3.0`, 5.0},
		{`7.0 % 3.0`, 1.0},
		{`2.0 ** 10.0`, 1024.0},
		{`-3.5`, -3.5},
		{`-(-3.5)`, 3.5},
	}
	runVmTests(t, tests)
}

func TestBigIntOperations(t *testing.T) {
	program := parse(`2 ** 100 + 1`)
	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	vm := New(comp.Bytecode())
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}
	obj := vm.LastPoppedStackElem()
	_, ok := obj.(*object.BigInteger)
	if !ok {
		t.Fatalf("expected BigInteger, got=%T", obj)
	}
}

func TestBooleanOperations(t *testing.T) {
	tests := []vmTestCase{
		{`true and true`, true},
		{`true and false`, false},
		{`false and true`, false},
		{`false and false`, false},
		{`true or false`, true},
		{`false or true`, true},
		{`false or false`, false},
		{`true or true`, true},
	}
	runVmTests(t, tests)
}

func TestBitwiseOperations(t *testing.T) {
	program := parse(`~0`)
	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	vm := New(comp.Bytecode())
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}
	obj := vm.LastPoppedStackElem()
	integer, ok := obj.(*object.Integer)
	if !ok {
		t.Fatalf("expected Integer, got=%T", obj)
	}
	if integer.Value != -1 {
		t.Fatalf("expected -1, got=%d", integer.Value)
	}
}

func TestRangeOperations(t *testing.T) {
	program := parse(`1..5`)
	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	vm := New(comp.Bytecode())
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}
	obj := vm.LastPoppedStackElem()
	_, ok := obj.(*object.List)
	if !ok {
		t.Fatalf("expected List, got=%T", obj)
	}
}

func TestRangeOrdering(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{`1..5`, []int{1, 2, 3, 4, 5}},
		{`5..1`, []int{5, 4, 3, 2, 1}},
		{`0..<4`, []int{0, 1, 2, 3}},
		{`4..<0`, []int{4, 3, 2, 1}},
		{`3..3`, []int{3}},
		{`3..<3`, []int{}},
	}
	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error for %s: %s", tt.input, err)
		}
		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error for %s: %s", tt.input, err)
		}
		testExpectedObject(t, tt.expected, vm.LastPoppedStackElem())
	}
}

// Regression test: applyFunctionFast used to run callbacks on a fixed-size
// frames array, so a callback that itself called another blue function
// panicked with an index-out-of-range. Callbacks must support full call
// depth like any other execution context.
func TestHigherOrderCallbackWithNestedCalls(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			fun g(x) { return x * 2; }
			fun f(x) { return g(x) + 1; }
			val l = [1, 2, 3].map(|x| => f(x));
			var total = 0;
			for v in l { total += v; }
			total
			`,
			expected: 15,
		},
		{
			input: `
			fun double(x) { return x * 2; }
			val l = [1, 2, 3];
			val mapped = l.map(|x| => double(x));
			mapped.filter(|x| => x > 2)[0]
			`,
			expected: 4,
		},
	}
	runVmTests(t, tests)
}

func TestAllAnyBuiltins(t *testing.T) {
	tests := []vmTestCase{
		{`all([true, true, true], |e| => e)`, true},
		{`all([true, false, true], |e| => e)`, false},
		{`all([], |e| => e)`, true},
		{`any([false, false, false], |e| => e)`, false},
		{`any([false, true, false], |e| => e)`, true},
		{`any([], |e| => e)`, false},
	}
	runVmTests(t, tests)
}

func TestMapFilterBuiltins(t *testing.T) {
	tests := []vmTestCase{
		{`map([1, 2, 3], |e| => e + 1)`, []int{2, 3, 4}},
		{`filter([1, 2, 3, 4, 5], |e| => e > 2)`, []int{3, 4, 5}},
	}
	runVmTests(t, tests)
}

func TestToNumBuiltin(t *testing.T) {
	tests := []vmTestCase{
		{`to_num("123")`, 123},
		{`to_num("-456")`, -456},
		{`to_num("0")`, 0},
	}
	runVmTests(t, tests)
}

func TestNegationOperations(t *testing.T) {
	tests := []vmTestCase{
		{`-5`, -5},
		{`-(-5)`, 5},
		{`--5`, 5},
		{`-0`, 0},
		{`-(-(-5))`, -5},
	}
	runVmTests(t, tests)
}

func TestComprehensions(t *testing.T) {
	tests := []vmTestCase{
		{`[x * 2 for x in [1, 2, 3]]`, []int{2, 4, 6}},
	}
	runVmTests(t, tests)
}

func TestIfElseExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`if (true) { 1 } else { 2 }`, 1},
		{`if (false) { 1 } else { 2 }`, 2},
		{`if (false) { 1 } else if (true) { 2 } else { 3 }`, 2},
	}
	runVmTests(t, tests)
}

func TestForLoops(t *testing.T) {
	tests := []vmTestCase{
		{`var sum = 0; for x in [1, 2, 3] { sum += x }; sum`, 6},
	}
	runVmTests(t, tests)
}

func TestErrorCases(t *testing.T) {
	tests := []vmTestCase{
		{`fun() { 1; }(1)`, `wrong number of arguments: want=0, got=1`},
		{`fun(a) { a; }()`, `wrong number of arguments: want=1, got=0`},
		{`fun(a, b) { a + b; }(1)`, `wrong number of arguments: want=2, got=1`},
	}
	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		vm := New(comp.Bytecode())
		err = vm.Run()
		if err == nil {
			t.Fatalf("expected VM error but resulted in none for: %s", tt.input)
		}
		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error: want=%q, got=%q", tt.expected, err.Error())
		}
	}
}

func TestTruthyFalsy(t *testing.T) {
	tests := []vmTestCase{
		{`if (1) { 1 } else { 2 }`, 1},
		{`if (null) { 1 } else { 2 }`, 2},
	}
	runVmTests(t, tests)
}

func TestListOperations(t *testing.T) {
	tests := []vmTestCase{
		{`[1, 2, 3] + [4, 5]`, []int{1, 2, 3, 4, 5}},
		{`[1] + []`, []int{1}},
		{`[] + [1]`, []int{1}},
		{`[1, 2, 3][0]`, 1},
		{`[1, 2, 3][1]`, 2},
		{`[1, 2, 3][2]`, 3},
		{`[[1, 2], [3, 4]][0][1]`, 2},
	}
	runVmTests(t, tests)
}

func TestSliceOperations(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Lists
		{`[1, 2, 3, 4, 5][1..3]`, `[2, 3]`},
		{`[1, 2, 3, 4, 5][1..<4]`, `[2, 3]`},
		{`[1, 2, 3, 4, 5][1..4]`, `[2, 3, 4]`},
		{`[1, 2, 3, 4, 5][0..5]`, `[1, 2, 3, 4, 5]`},
		{`[1, 2, 3, 4, 5][0..0]`, `[]`},
		{`[1, 2, 3, 4, 5][5..5]`, `[]`},
		{`[1, 2, 3][3..3]`, `[]`},
		{`[1, 2, 3][2..1]`, `[]`},
		// Stepped slices
		{`[1, 2, 3, 4, 5, 6, 7, 8, 9, 10][0..10..3]`, `[1, 4, 7, 10]`},
		{`[1, 2, 3, 4, 5, 6, 7, 8, 9, 10][1..9..2]`, `[2, 4, 6, 8]`},
		{`[1, 2, 3, 4, 5][0..<5..2]`, `[1, 3]`},
		{`val r = 1..20; r[0..<20..2]`, `[1, 3, 5, 7, 9, 11, 13, 15, 17, 19]`},
		// Strings (indexed by rune, not byte)
		{`'abcdef'[1..3]`, `bc`},
		{`'abcdef'[1..<4]`, `bc`},
		{`'abcdef'[1..4]`, `bcd`},
		{`'abcdef'[0..6]`, `abcdef`},
		{`'abcdef'[0..0]`, ``},
		{`'abcdef'[6..6]`, ``},
		{`'abcdef'[0..6..2]`, `ace`},
		{`'abcdef'[1..6..3]`, `be`},
		{`'héllo'[0..2]`, `hé`},
		{`'héllo'[1..3]`, `él`},
		{`'héllo'[0..5..2]`, `hlo`},
		// Sets preserve insertion order
		{`{1, 2, 3, 4, 5}[1..3]`, `{2, 3}`},
		{`{1, 2, 3, 4, 5}[1..<4]`, `{2, 3}`},
		{`{1, 2, 3, 4, 5, 6}[0..6..2]`, `{1, 3, 5}`},
		{`{1, 2, 3}[3..3]`, `{}`},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSliceErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Out-of-bounds start/stop
		{`[1, 2, 3][0..10]`, `slice index out of bounds. 0, 10 for len 3`},
		{`[1, 2, 3][-1..2]`, `slice index out of bounds. -1, 2 for len 3`},
		{`'abc'[0..5]`, `slice index out of bounds. 0, 5 for len 3`},
		{`'abc'[-1..2]`, `slice index out of bounds. -1, 2 for len 3`},
		{`{1, 2, 3}[0..10]`, `slice index out of bounds. 0, 10 for len 3`},
		// Non-integer bounds are errors, not Go interface-conversion panics
		{`[1, 2, 3][1.5..2]`, `slice cannot be created with start index type: FLOAT`},
		{`[1, 2, 3][0..2.5]`, `slice cannot be created with stop index type: FLOAT`},
		{`[1, 2, 3][[1]..2]`, `slice cannot be created with start index type: LIST`},
		{`[1, 2, 3][0..[2]]`, `slice cannot be created with stop index type: LIST`},
		// Bad steps are errors, not a hang (0) or a Go index panic (negative)
		{`[1, 2, 3][0..2..0]`, `slice cannot be created with a step size of 0`},
		{`[1, 2, 3][0..2..-1]`, `slice cannot be created with a step size of -1`},
		{`[1, 2, 3][0..2..1.5]`, `slice cannot be created with step size type: FLOAT`},
		// Only list, set and string are sliceable
		{`5[0..1]`, `slice cannot be created with type: INTEGER`},
	}
	for _, tt := range tests {
		runExpectVmError(t, tt.input, tt.want)
	}
}

func TestDefaultArgs(t *testing.T) {
	tests := []vmTestCase{
		{`fun foo(a=10) { a }; foo()`, 10},
		{`fun foo(a=10) { a }; foo(5)`, 5},
		{`fun foo(a=1, b=2, c=3) { a + b + c }; foo()`, 6},
		{`fun foo(a=1, b=2, c=3) { a + b + c }; foo(c=10)`, 13},
	}
	runVmTests(t, tests)
}

func TestGlobalAndLocalVars(t *testing.T) {
	tests := []vmTestCase{
		{`var x = 10; x`, 10},
		{`val x = 10; x`, 10},
	}
	runVmTests(t, tests)
}

func TestReturnStatements(t *testing.T) {
	tests := []vmTestCase{
		{`fun f() { return 42; }; f()`, 42},
	}
	runVmTests(t, tests)
}

func TestLambdaFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`var f = |x, y| => x + y; f(3, 4)`, 7},
	}
	runVmTests(t, tests)
}

func TestTypeChecking(t *testing.T) {
	tests := []vmTestCase{
		{`type(1)`, ""},
	}
	runVmTests(t, tests)
}

func testUIntegerObject(expected uint64, actual object.Object) error {
	result, ok := actual.(*object.UInteger)
	if !ok {
		return fmt.Errorf("object is not UInteger. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}
	return nil
}

// Regression: binaryUIntegerOp never assigned rightVal when the right operand
// was an Integer, so every mixed uint/int op silently computed against 0.
func TestMixedUIntegerIntegerOperations(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"0u10 + 5", 15},
		{"5 + 0u10", 15},
		{"0u10 - 4", 6},
		{"0u10 * 3", 30},
		{"0u9 / 2", 4},
		{"0u10 % 3", 1},
		{"var u = 0u10; u += 5; u", 15},
		{"var u = 0u10; u *= 3; u", 30},
	}
	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			t.Fatalf("compiler error in %q: %s", tt.input, err)
		}
		vm := New(comp.Bytecode())
		if err := vm.Run(); err != nil {
			t.Fatalf("vm error in %q: %s", tt.input, err)
		}
		if err := testUIntegerObject(tt.expected, vm.LastPoppedStackElem()); err != nil {
			t.Errorf("%q: %s", tt.input, err)
		}
	}
}

func TestMixedUIntegerComparisons(t *testing.T) {
	tests := []vmTestCase{
		{"0u10 == 10", true},
		{"10 == 0u10", true},
		{"0u10 != 10", false},
		{"0u10 > 4", true},
		{"0u10 >= 10", true},
	}
	runVmTests(t, tests)
}

func TestMixedUIntegerNegativeRightErrors(t *testing.T) {
	program := parse(`0u10 + (-5)`)
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	vm := New(comp.Bytecode())
	err := vm.Run()
	if err == nil {
		t.Fatal("expected vm error for negative right Integer in uint op")
	}
	want := "Right Integer was negative, and is not allowed for Unsigned Integer operations."
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q does not contain %q", err.Error(), want)
	}
}

func TestSetInstructionOffsetRunsOnlyNewInstructions(t *testing.T) {
	comp := compiler.New()
	if err := comp.Compile(parse("var a = 99")); err != nil {
		t.Fatalf("compile first chunk: %s", err)
	}
	firstLen := len(comp.Bytecode().Instructions)

	if err := comp.Compile(parse("1 + 2")); err != nil {
		t.Fatalf("compile second chunk: %s", err)
	}
	bc := comp.Bytecode()
	if len(bc.Instructions) <= firstLen {
		t.Fatalf("second chunk added no instructions")
	}

	vm := New(bc)
	vm.SetInstructionOffset(firstLen)
	if err := vm.Run(); err != nil {
		t.Fatalf("run: %s", err)
	}
	got := vm.LastPoppedStackElem()
	if got == nil || got.Inspect() != "3" {
		t.Fatalf("last popped = %v, want 3", got)
	}
	for i, g := range vm.globals {
		if g != nil {
			t.Fatalf("global %d = %v, first chunk should not have run", i, g)
		}
	}
}

func TestTensorSliceOperations(t *testing.T) {
	const x = `import ml; val x = ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0], [7.0, 8.0, 9.0]]); `
	tests := []struct {
		input string
		want  string
	}{
		{x + `ml.equal(x[0..2], ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]))`, "true"},
		{x + `ml.equal(x[0..<3], ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]))`, "true"},
		{x + `ml.equal(x[0..3..2], ml.tensor([[1.0, 2.0, 3.0], [7.0, 8.0, 9.0]]))`, "true"},
		{x + `ml.equal(x[-1..3], ml.tensor([[7.0, 8.0, 9.0]]))`, "true"},
		{x + `ml.equal(x[0..-1], ml.tensor([[1.0, 2.0, 3.0], [4.0, 5.0, 6.0]]))`, "true"},
		{x + `ml.equal(x[0..100], x)`, "true"},
		{x + `ml.equal(x[2..-1..-1], ml.tensor([[7.0, 8.0, 9.0], [4.0, 5.0, 6.0], [1.0, 2.0, 3.0]]))`, "true"},
		{x + `ml.equal(x[1], ml.tensor([4.0, 5.0, 6.0]))`, "true"},
		{x + `ml.equal(x[-1], ml.tensor([7.0, 8.0, 9.0]))`, "true"},
		{x + `x[1][2].item()`, "6.0"},
		{`import ml; ml.equal(ml.tensor([1.0, 2.0, 3.0, 4.0, 5.0])[1..3], ml.tensor([2.0, 3.0]))`, "true"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTensorSliceErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`import ml; ml.tensor([1.0, 2.0, 3.0])[0..3..0]`, `step size of 0`},
		{`import ml; ml.tensor([1.0, 2.0, 3.0])[1.5..2]`, `slice cannot be created with start index type: FLOAT`},
		{`import ml; ml.tensor([1.0, 2.0, 3.0])[0..2.5]`, `slice cannot be created with stop index type: FLOAT`},
	}
	for _, tt := range tests {
		runExpectVmError(t, tt.input, tt.want)
	}
}

func TestSendFunctionToProcess(t *testing.T) {
	got := runInspect(t, `
fun worker(parent) {
    val msg = self().recv();
    parent.send("got " + msg);
}
val me = self();
val p = spawn(worker, [me]);
send(p, "hi");
me.recv()
`)
	if got != "got hi" {
		t.Errorf("send(pid, msg) result = %q, want %q", got, "got hi")
	}
}

func TestRecvFunctionFromProcess(t *testing.T) {
	got := runInspect(t, `
fun worker() {
    self().send("hello from worker");
}
val p = spawn(worker, []);
recv(p)
`)
	if got != "hello from worker" {
		t.Errorf("recv(pid) result = %q, want %q", got, "hello from worker")
	}
}

func TestProcessSendUsesSharedHelp(t *testing.T) {
	v := New(&bluec.Bytecode{})
	process := &object.Process{Ch: make(chan object.Object, 1)}
	if err := v.executeProcessIndexExpression(process, "send"); err != nil {
		t.Fatalf("executeProcessIndexExpression: %s", err)
	}
	send, ok := v.pop().(*object.Builtin)
	if !ok {
		t.Fatalf("expected *object.Builtin, got %T", v.pop())
	}
	if send.HelpStr != object.SendHelpString() {
		t.Errorf("process send help does not match shared send help")
	}
}

func TestProcessRecvUsesSharedHelp(t *testing.T) {
	v := New(&bluec.Bytecode{})
	process := &object.Process{Ch: make(chan object.Object, 1)}
	if err := v.executeProcessIndexExpression(process, "recv"); err != nil {
		t.Fatalf("executeProcessIndexExpression: %s", err)
	}
	recv, ok := v.pop().(*object.Builtin)
	if !ok {
		t.Fatalf("expected *object.Builtin, got %T", v.pop())
	}
	if recv.HelpStr != object.RecvHelpString() {
		t.Errorf("process recv help does not match shared recv help")
	}
}

func TestNullEqualityShortCircuits(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`null == null`, "true"},
		{`null != null`, "false"},

		{`val m = {"a": 1}; m == null`, "false"},
		{`val m = {"a": 1}; null == m`, "false"},
		{`val m = {"a": 1}; m != null`, "true"},
		{`val m = {"a": 1}; null != m`, "true"},
		{`val m = {}; m == null`, "false"},

		{`val l = [1, 2]; l == null`, "false"},
		{`val l = [1, 2]; null == l`, "false"},
		{`val l = [1, 2]; l != null`, "true"},

		{`val s = {1, 2}; s == null`, "false"},
		{`val s = {1, 2}; null == s`, "false"},

		{`val st = @{x: 1}; st == null`, "false"},
		{`val st = @{x: 1}; null == st`, "false"},
		{`val st = @{x: 1}; st != null`, "true"},

		{`0 == null`, "false"},
		{`null == 0`, "false"},
		{`1.5 == null`, "false"},
		{`"" == null`, "false"},
		{`false == null`, "false"},
		{`null == false`, "false"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEqualityIdentityShortCircuits(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`val m = {"a": 1}; m == m`, "true"},
		{`val m = {"a": 1}; m != m`, "false"},
		{`var m = {"a": 1}; m["b"] = 2; m == m`, "true"},
		{`val l = [1, [2, 3]]; l == l`, "true"},
		{`val l = [1, [2, 3]]; l != l`, "false"},
		{`val s = {1, 2}; s == s`, "true"},
		{`val st = @{x: 1}; st == st`, "true"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStructuralEqualityStillWorks(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`val a = {"x": 1}; val b = {"x": 1}; a == b`, "true"},
		{`val a = {"x": 1}; val b = {"x": 2}; a == b`, "false"},
		{`val a = {"x": 1}; val b = {"x": 2}; a != b`, "true"},
		{`val a = [1, [2, 3]]; val b = [1, [2, 3]]; a == b`, "true"},
		{`val a = [1, [2, 3]]; val b = [1, [2, 4]]; a == b`, "false"},
		{`val a = [1, [2, 3]]; val b = [1, [2, 4]]; a != b`, "true"},
		{`val a = {1, 2}; val b = {1, 2}; a == b`, "true"},
		{`val a = {1, 2}; val b = {1, 3}; a == b`, "false"},
		{`val a = @{x: 1}; val b = @{x: 1}; a == b`, "true"},
		{`val a = @{x: 1}; val b = @{x: 2}; a == b`, "false"},
	}
	for _, tt := range tests {
		if got := runInspect(t, tt.input); got != tt.want {
			t.Errorf("%s = %q, want %q", tt.input, got, tt.want)
		}
	}
}
