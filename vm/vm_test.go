package vm

import (
	"blue/compiler"
	"blue/object"
	"fmt"
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
	for _, tt := range tests {
		program := parse(tt.input)
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
		{"[1][-1]", object.NULL},
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
		{
			`len(1)`,
			&object.Error{
				Message: "PositionalTypeError: `len` expects argument 1 to be STRING, LIST, MAP, SET, or BYTES. got=INTEGER",
			},
		},
		{`len("one", "two")`,
			&object.Error{
				Message: "InvalidArgCountError: `len` wrong number of args. got=2, want=1",
			},
		},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`print("hello", "world!")`, object.NULL},
		{`push([], 1)`, 1},
		{`push(1, 1)`, &object.Error{
			Message: "PositionalTypeError: `push` expects argument 1 to be LIST. got=INTEGER",
		},
		},
	}
	runVmTests(t, tests)
}
