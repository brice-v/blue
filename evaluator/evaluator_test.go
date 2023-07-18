package evaluator

import (
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"log"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"1_000", 1000},
		{"1_000_000", 1000000},
		{"-5", -5},
		{"-10", -10},
		{"-1_000", -1000},
		{"-1_000_000", -1000000},
		{"5_00 + 5_00 + 5_0_0 + 500 - 1_000", 1000},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-5_0 + 10_0 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"2_0 + 2 * -1_0", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"2 ** 3", 8},
		{"100 // 3", 33},
		{"100 % 3", 1},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalHexExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"0x05", 0x05},
		{"0x10", 0x10},
		{"0x10_00_A", 0x1000A},
		{"0x10_00_000AB", 0x1000000AB},
		{"0xFF - 0xF0", 0x0F},
		{"0x1234 + 0x1234", 0x2468},
		{"0x0001 * 0x0002", 0x02},
		{"0x01 - 0x02", 0xFFFFFFFFFFFFFFFF},
		{"0xFF / 0xf0", uint64(0xFF / 0xF0)},
		{"0xF0F0 & 0x0F0F", 0x0},
		{"0xF0F0 | 0xF0F0", 0xF0F0},
		{"0xF0F0 ^ 0xF0F0", 0x0},
		{"0x0002 ** 0x000F", 0x8000},
		{"0x0001 << 0xF", 0x8000},
		{"0xF000 >> 0xF", 0x1},
		{"~0x0F0F", 0xFFFFFFFFFFFFF0F0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testUintObject(t, evaluated, tt.expected)
	}
}

func TestEvalOctalExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"0o05", 005},
		{"0o10", 010},
		{"0o10_00_7", 010007},
		{"0o10_00_00067", 0100000067},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testUintObject(t, evaluated, tt.expected)
	}
}
func TestEvalBinExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"0b11", 3},
		{"0b10", 2},
		{"0b10_00_111", 71},
		{"0b10_00_00011", 259},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testUintObject(t, evaluated, tt.expected)
	}
}

func TestEvalFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"123.14_244", 123.14244},
		{"0.1_2_3_4_5", 0.12345},
		{"1_000_000.999", 1000000.999},
		{"1234.53", 1234.53},
		{"-123.14_244", -123.14244},
		{"-0.1_2_3_4_5", -0.12345},
		{"-1_000_000.999", -1000000.999},
		{"-1234.53", -1234.53},
		{"-100.001 + 10_1.001", 1.0},
		{"5.0 + 2.0 * 10.01", 25.02},
		{"2_0.00 + 2.0 * -1_0.0", 0.0},
		{"50.0 / 2.0 * 2.0 - 100.5", -50.5},
		{"2.0 * (5.0 + 10.0)", 30.0},
		{"3.0 * 3.0 * 3.0 + 10.0", 37.0},
		{"3.0 * (3.0 * 3.0) + 10.0", 37.0},
		{"(5.0 + 10.0 * 2.0 + 15.0 / 3.0) * 2.0 + -10.0", 50.0},
		{"2.0 ** 3.0", 8.0},
		{"100.0 // 3.0", 33.0},
		{"100.0 % 3.0", 1.0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testFloatObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 <= 2", true},
		{"1 >= 2", false},
		{"1 <= 1", true},
		{"1 >= 1", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"(1 <= 2) == true", true},
		{"(1 <= 2) == false", false},
		{"(1 >= 2) == true", false},
		{"(1 >= 2) == false", true},
		{"(true and false) == false", true},
		{"(true and false) != false", false},
		{"(false or true) != false", true},
		{"((1 < 2) and (2 >= 1)) == true", true},
		{"((1 < 2) or (2 >= 1)) == false", false},
		{"((1 > 2) and true) == true", false},
		{"((1 > 2) or true) == false", false},
		{"((1 <= 2) and not false) == true", true},
		{"(1 <= 2) and not true) == false", false},
		{"((1 >= 2) and false) == true", false},
		{"((1 >= 2) or false) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		if !testBooleanObject(t, evaluated, tt.expected) {
			log.Printf("failed on tt.input = %s\n", tt.input)
		}
	}
}

func TestEvalBooleanWithFloatExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1.0 < 2.0", true},
		{"1.0_00 > 2.0_00", false},
		{"1.0 < 1.0", false},
		{"1.01 > 1.01", false},
		{"1.123 <= 2.234", true},
		{"1.123 >= 2.12_3", false},
		{"1.1 <= 1.1", true},
		{"1.1 >= 1.1", true},
		{"(1.1 < 2.2) == true", true},
		{"(1.1 < 2.2) == false", false},
		{"(1.1 > 2.2) == true", false},
		{"(1.1 > 2.2) == false", true},
		{"(1.23 <= 2.23) == true", true},
		{"(1.23 <= 2.23) == false", false},
		{"(1.23_4_5 >= 2.23_45) == true", false},
		{"(1.001 >= 2.001) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{`
		if (10 > 1) {
			if (10 > 1) {
				return 10;
			}
			return 1;
		}`, 10},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
			if (10 > 1) {
				if (10 > 1) {
					return true + false;
				}
				return 1;
			}
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`{"name": "Monkey"}[fun(x) { x }];`,
			"unusable as a map key: FUNCTION",
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)", evaluated, evaluated)
			continue
		}
		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expectedMessage, errObj.Message)
		}
	}
}

func TestValStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"val a = 5; a;", 5},
		{"val a = 5 * 5; a;", 25},
		{"val a = 5; val b = a; b;", 5},
		{"val a = 5; val b = a; val c = a + b + 5; c;", 15},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestVarStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var a = 5; a;", 5},
		{"var a = 5 * 5; a;", 25},
		{"var a = 5; var b = a; b;", 5},
		{"var a = 5; var b = a; var c = a + b + 5; c;", 15},
		{"var a = 5; a = 10; a += 10; a;", 20},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fun(x) { x + 2; };"
	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}
	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}
	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionStatement(t *testing.T) {
	input := "fun abc(x) { x + 2 };  abc(4);"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 6)
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"val identity = fun(x) { x; }; identity(5);", 5},
		{"val identity = fun(x) { return x; }; identity(5);", 5},
		{"val double = fun(x) { x * 2; }; double(5);", 10},
		{"val add = fun(x, y) { x + y; }; add(5, 5);", 10},
		{"val add = fun(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fun(x) { x; }(5)", 5},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestStringExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"This is a string."`, "This is a string."},
		{`"Hello World!"`, "Hello World!"},
		{`var x = 5 + 3; "Testing #{x}"`, "Testing 8"},
		{`"Testing #{9 ** 3}"`, "Testing 729"},
		{`"Testing #{  2 ** (3 + 5)}  #{2.0 + 3.0}"`, "Testing 256  5.000000"},
		{`"Test #{}"`, "Test "},
		{`var x = "Hey Another String" "Testing #{9 ** 3}" + " #{x}"`, "Testing 729 Hey Another String"},
	}
	for _, tt := range tests {
		testStringObject(t, testEval(tt.input), tt.expected)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "PositionalTypeError: `len` expects argument 1 to be STRING, LIST, MAP, SET, or BYTES. got=INTEGER"},
		{`len("one", "two")`, "InvalidArgCountError: `len` wrong number of args. got=2, want=1"},
		{`len([1,2,3,4,5])`, 5},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"
	evaluated := testEval(input)
	result, ok := evaluated.(*object.List)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}
	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}
	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"val i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"val myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"val myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"val myArray = [1, 2, 3]; val i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			&object.Error{Message: "index out of bounds: length=3, index=-1"},
		},
		{
			"val myArray = [1, 2, 3]; val i = 1; myArray.i",
			2,
		},
		{
			"val myArray = [1, 2, 3]; val i = myArray.1; myArray.i",
			3,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch obj := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(obj))
		case *object.Error:
			testErrorObject(t, evaluated, tt.expected.(*object.Error).Message)
		default:
			testNullObject(t, evaluated)
		}
	}
}

func TestMapLiterals(t *testing.T) {
	input := `val two = "two";
	{
		one: 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`
	evaluated := testEval(input)
	result, ok := evaluated.(*object.Map)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}
	expected := map[object.HashKey]int64{
		(&object.Stringo{Value: "one"}).HashKey():   1,
		(&object.Stringo{Value: "two"}).HashKey():   2,
		(&object.Stringo{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():       4,
		TRUE.HashKey():                              5,
		FALSE.HashKey():                             6,
	}

	if result.Pairs.Len() != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", result.Pairs.Len())
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs.Get(expectedKey)
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}
		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestMapIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`val key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}.bar`,
			nil,
		},
		{
			`{"foo": 5}.foo`,
			5,
		},
		{
			`{}.foo`,
			nil,
		},
		{
			`val x = {name: "brice", nested: {other: 10}}; x.nested.other`,
			10,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestForExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`var x = 0; for (i in 1..10) { x+=1; if (i == 5) { break; error("UNREACHABLE"); } }; x`,
			5,
		},
		{
			`var x = 0; for (i in 1..10) { x+=1; if (i == 5) { break; error("UNREACHABLE"); x += 10;} }; x`,
			5,
		},
		{
			`var i = 0; for (true) { for (x in 1..10) { if (i > 30) { break; error("UNREACHABLE"); i += 100; } i += 1; }; i += 1; if (i < 100) { continue; error("UNREACHABLE"); } else { i += 2000; break; error("UNREACHABLE"); i += 300; } }; i`,
			2100,
		},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

// Helper functions

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func testEval(input string) object.Object {
	l := lexer.New(input, "<internal: test>")
	p := parser.New(l)
	program := p.ParseProgram()
	e := New()
	return e.Eval(program)
}

// testIntegerObject tests to see if the given object is an integer and if its
// value matches the expected value, returns true if so
func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("obj is not *object.Integer. got=%T", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d want=%d", result.Value, expected)
		return false
	}

	return true
}

func testErrorObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("obj is not *object.Error. got=%T", obj)
		return false
	}

	if result.Message != expected {
		t.Errorf("object has wrong value. got=%s want=%s", result.Message, expected)
		return false
	}

	return true
}

// testUintObject tests if the given object is a Uint and if it matches the given Uint
func testUintObject(t *testing.T, obj object.Object, expected uint64) bool {
	result, ok := obj.(*object.UInteger)
	if !ok {
		t.Errorf("obj is not *object.UInteger. got=%T", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%x want=%x", result.Value, expected)
		return false
	}

	return true
}

// testFloatObject checks if the given object is a float and if it matches the given expected float
func testFloatObject(t *testing.T, obj object.Object, expected float64) bool {
	result, ok := obj.(*object.Float)
	if !ok {
		t.Errorf("obj is not *object.Float. got=%T", obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%f want=%f", result.Value, expected)
		return false
	}

	return true
}

// testBooleanObject checks to see if the given boolean value matches the evaluated and returns true if so
func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("obj is not *object.Boolean. got=%T", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t want=%t", result.Value, expected)
		return false
	}

	return true
}

// testStringObject tests to see if the given object is a string and if its
// value matches the expected value, returns true if so
func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.Stringo)
	if !ok {
		t.Errorf("obj is not *object.Stringo. got=%T", obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%s want=%s", result.Value, expected)
		return false
	}

	return true
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"not true", false},
		{"not false", true},
		{"not 5", false},
		{"not not true", true},
		{"not not false", false},
		{"not not 5", true},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestGoObjectToBlueObject(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"true"},
		{"[1,2,3]"},
		{"{'1': 'hello'}"},
		{"{'1': [1,2,{'a':'b'}]}"},
	}
	for i, tt := range tests {
		evaluated := testEval(tt.input)
		goObj, err := blueObjectToGoObject(evaluated)
		if err != nil {
			t.Fatalf("[%d] %s", i, err.Error())
		}
		blueObj, err := goObjectToBlueObject(goObj)
		if err != nil {
			t.Fatalf("[%d] %s", i, err.Error())
		}
		if object.HashObject(evaluated) != object.HashObject(blueObj) {
			t.Fatalf("[%d] initial evaluated blue obj did not match goObj back to blueObj: %s", i, tt.input)
		}
	}
}
