package b_program_test

import (
	"blue/ast"
	"blue/blueutil"
	"blue/compiler"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/vm"
	"bytes"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testDirectoryWithVm(t *testing.T, path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}

	// Disable caching that breaks tests
	blueutil.ENABLE_VM_CACHING = false

	for _, f := range files {
		// test_http is still not setup to work yet
		if f.Name() == "test_http.b" {
			continue
		}
		executeBlueTestFileWithVm(path, f, t)
	}
}

func TestAllProgramsInDirectoryWithVm(t *testing.T) {
	testDirectoryWithVm(t, "./")
}

func TestAllGeneratedProgramsInDirectoryWithVm(t *testing.T) {
	testDirectoryWithVm(t, "./generated")
}

func executeBlueTestFileWithVm(dir string, f fs.DirEntry, t *testing.T) {
	if !strings.HasSuffix(f.Name(), ".b") {
		return
	}
	fpath := filepath.Join(dir, f.Name())
	// Note: Comment out this defered func to see what the panic trace is
	// defer func() {
	// 	// recover from panic if one occured. Set err to nil otherwise.
	// 	err := recover()
	// 	if err != nil {
	// 		t.Fatalf("PANIC in FILE %s Error: %+v", fpath, err)
	// 	}
	// }()
	openFile, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = openFile.Close()
		if err != nil {
			log.Printf("Failed to close file with path: %s, error: %s", fpath, err.Error())
		}
	}()

	data, err := io.ReadAll(openFile)
	if err != nil {
		t.Fatal(err)
	}
	stringData := string(data)
	if strings.HasPrefix(stringData, "# IGNORE") || strings.HasPrefix(stringData, "#IGNORE") {
		return
	}
	if strings.HasPrefix(stringData, "#VM IGNORE") || strings.HasPrefix(stringData, "# VM IGNORE") {
		return
	}
	l := lexer.New(stringData, fpath)

	p := parser.New(l)
	program := p.ParseProgram()
	if p.HasErrors() {
		p.PrintParserErrors(os.Stderr)
		t.Fatalf("File `%s`: failed to parse", f.Name())
	}
	globals := make([]object.Object, vm.GlobalsSize)
	c := compiler.NewFromCore()
	err = c.Compile(program)
	if err != nil {
		t.Errorf("File `%s`: compiler returned error %s", f.Name(), err.Error())
		return
	}
	v := vm.NewWithGlobalsStore(c.Bytecode(), globals)
	err = v.Run()
	if err != nil {
		t.Errorf("File `%s`: vm returned error %s", f.Name(), err.Error())
		return
	}
	obj := v.LastPoppedStackElem()
	if obj.Type() == object.ERROR_OBJ {
		errorObj := obj.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(errorObj.Message)
		buf.WriteByte('\n')
		// for e.ErrorTokens.Len() > 0 {
		// 	buf.WriteString(lexer.GetErrorLineMessage(e.ErrorTokens.PopBack()))
		// 	buf.WriteByte('\n')
		// }
		t.Errorf("File `%s`: vm returned error: %s", f.Name(), buf.String())
	}
	// TODO: look into why lastPoppedStackElem is not true with asserts
	// if obj.Inspect() != "true" {
	// 	t.Errorf("File `%s`: Did not return true as last statement. Failed", f.Name())
	// }
	object.ClearGlobalState()
}

func TestVmStackOverflowForIn(t *testing.T) {
	s := `fun main() {
		for i in 1..5000 {
			println("Hello World #{i}!");
		}
	}

	main();`
	vmStringWithCore(t, s)
}

func TestVmNotEqualIssue(t *testing.T) {
	s := `var abc = 1;
	assert(abc != 5);`
	vmString(t, s)
}

func TestVmArgCountIssue(t *testing.T) {
	s := `fun random_fun(a) {
		return a;
	}

	fun other(a, b, c, d=true) {
		"#{a}#{b}#{c}#{d}"
	}

	val result = random_fun('TEST').other('R','E');
	assert(result == "TESTREtrue");`
	vmString(t, s)
}

func TestVmArgCountIssue2(t *testing.T) {
	s := `fun random_fun(a) {
		return a;
	}

	fun other(a, b, c, d=true) {
		"#{a}#{b}#{c}#{d}"
	}

	assert('test'.random_fun().other('R','E') == 'testREtrue');`
	vmString(t, s)
}

func TestBrokenLongCall(t *testing.T) {
	s := `'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
	val line = "11A = (11B, XXX)"
	val values = line.split(" = ")[1].split("");
	assert("#{values}" == "[(, 1, 1, B, ,,  , X, X, X, )]")`
	vmStringWithCore(t, s)
}

func TestBrokenReturnForAll(t *testing.T) {
	s := `val y = {'a': 1, 'b': 1};
	assert(y.values().all(|e| => e == 1));`
	vmStringWithCore(t, s)
}

func TestVmFixForLoop(t *testing.T) {
	s := `var input = """LR

	11A = (11B, XXX)
	11B = (XXX, 11Z)""".replace("\r", "");

	val split_lines = input.split("\n");
	var m = {};

	for var i = 2; i < split_lines.len(); i += 1 {
		val line = split_lines[i];
	}`
	vmStringWithCore(t, s)
}

func TestBrokenMapIndexing(t *testing.T) {
	s := `fun hello() {
		var this = {};
		var starting_keys = ['11A', '22A'];
		this['starting_keys'] = starting_keys;
		return this;
	}

	var game = hello();
	assert(game.starting_keys == ['11A','22A']);`
	vmStringWithCore(t, s)
}

func TestBrokenMapIndexing2(t *testing.T) {
	s := `fun hello() {
		var this = {};
		var a = 2;
		this['a'] = a;
		this.next = fun() {
			this.a + 1
		}
		return this;
	}

	var game = hello();
	assert(game.next() == 3);`
	vmStringWithCore(t, s)
}

func TestIncorrectTypeInAdd(t *testing.T) {
	s := `var x = 2 ** 100;
	assert(x.type() == 'BIG_INTEGER');

	var y = x + 0.5;
	assert(y.type() == 'BIG_FLOAT');`
	vmStringWithCore(t, s)
}

func TestBrokenVmConfig(t *testing.T) {
	s := `import config
	var path_prefix = "b_test_programs/";
	try {
		config.load_file("#{path_prefix}test.yml")
	} catch (e) {
		path_prefix = ""
	}
	config.load_file("#{path_prefix}test.yml");`
	vmStringWithCore(t, s)
}

func TestBrokenVmRe(t *testing.T) {
	s := `var x = r/abc[\t|\s]/;
	var xx = re("abc[\\t|\\s]");`
	vmStringWithCore(t, s)
}

func TestBrokenVmTypeCallOnComprehension(t *testing.T) {
	s := `val t12 = type([x for (x in 1..2)]);`
	vmStringWithCore(t, s)

	s1 := `val t13 = type({x for (x in 1..10)});
	val t14 = type({x: "" for (x in 1..10)});`
	vmStringWithCore(t, s1)
}

func TestBrokenVmBuiltinAsMapKey(t *testing.T) {
	s := `var x = {wait: 'hello'}
	println(x['wait'])
	assert(x['wait'] == 'hello');`
	vmStringWithCore(t, s)
}

func TestBrokenVmSlices(t *testing.T) {
	s := `val x = [1,2,3,4,5,6];
	val y = x[1..3];
	val yy = x[1..<4];
	val z = [2,3,4];
	assert(y == z);
	assert(yy == z);`
	vmStringWithCore(t, s)
}

func TestBrokenVmIfShortcut(t *testing.T) {
	s := `'AAAAAAAAAAAAAAAAAAAAAAAAAAAA';
	var x = null;
	var y = fun() { assert(false); };
	if (x != null && y()) {
		println("SHOULD NOT PRINT")
		assert(false);
	}
	if (x == null || y()) {
		println("SHOULD PRINT")
		assert(true);
	}`
	vmStringWithCore(t, s)
}

func TestBrokenVmDefaultArgs(t *testing.T) {
	s := `fun hello(arg1=10, arg2="", arg3=true) {
		return "arg1 = #{arg1}, arg2 = #{arg2}, arg3 = #{arg3}";
	}
	val result = hello(arg1=100, "something");
	val expected = "arg1 = 100, arg2 = something, arg3 = true";
	assert(result == expected);`
	vmStringWithCore(t, s)
}

func TestBrokenVmOpOrWithBuiltin(t *testing.T) {
	s := `val i = 0; val lines = [];
	val x = i > len(lines) or (i+1 == 0);`
	vmStringWithCore(t, s)
}

func TestBrokenVmNullCoalescing(t *testing.T) {
	s := `val a = "Something";
	val b = null;
	val c = b or a;

	assert(c == a);`
	vmStringWithCore(t, s)
}

func TestBrokenVmOpAssign(t *testing.T) {
	s := `val a = "A";
	val b = "B";
	var c = "";
	'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA'
	c += a;
	c += b;
	assert(c == "AB");
	
	var x = {y: 'A', z: 'B', cc: ''};
	x.cc += x.y;
	x.cc += x.z;
	assert(x.cc == "AB");`
	vmStringWithCore(t, s)
}

func TestBrokenLogicInFetchVm(t *testing.T) {
	s := `val _f = fun(resource, method, headers, body, full_resp) {
		"#{resource}, #{method}, #{headers}, #{body}, #{full_resp}"
	}

	fun f(resource, options=null, full_resp=true) {
		if (options == null) {
			options = {
				method: 'GET',
				headers: {},
				body: null,
			};
		} else {
			val t = options.type();
			if (t != 'MAP') {
				return error("'fetch' error:  options must be MAP. got=#{t}");
			}
			if (options.method == null) {
				options.method = 'GET';
			}
			if (options.headers == null) {
				options.headers = {};
			} else {
				val ht = type(options.headers);
				if (ht != 'MAP') {
					return error("'fetch' error:  options.headers must be MAP. got=#{ht}");
				}
			}
			'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA';
			if (options.method == 'GET' or options.method == 'DELETE') {
				if (options.body != null) {
					return error("'fetch' error: options.body must be NULL for 'GET' or 'DELETE' methods");
				}
			}
		}
		_f(resource, options.method, options.headers, options.body, full_resp)
	}

	var result = f("http://localhost:3001/post/abc/213", {method: 'POST', body: '{"name":"john","pass":"doe"}', headers: {'content-type': 'application/json'}});
	assert(result == 'http://localhost:3001/post/abc/213, POST, {content-type: application/json}, {"name":"john","pass":"doe"}, true')`
	vmStringWithCore(t, s)

}

func TestBrokenVmStackOverflowIssue(t *testing.T) {
	s := `val input = """[[1],[2,3,4]]
	[[1],4]""";
	var lines = input.split("\n");

	fun get_pairs_of_lists(lines) {
		var pairs = [];
		var index = 1;
		var i = 0;
		for (true) {
			if (i > len(lines) or (i+1) > len(lines)) {
				break;
			}
			if (lines[i] == '') {
				i += 1;
			} else {
				var pair1 = eval(lines[i]);
				var pair2 = eval(lines[i+1]);
				pairs << {'pair1': pair1, 'pair2': pair2, 'index': index};
				index += 1;
				i += 2;
			}
		}
		return pairs;
	}

	fun part1(lines) {
		val pairs_of_lists = get_pairs_of_lists(lines);
		for (pair in pairs_of_lists) {
			println("!!!!IN HERE-------------------------------")
			println("pair = #{pair}")
		}
	}

	part1(lines);

	assert(true);`
	vmStringWithCore(t, s)
}

func TestAssignIndexWithBuiltinDotCallVm(t *testing.T) {
	s := `var this = {};
	this.println = "Hello";
	println(this.println);
	this.println.println();`
	vmStringWithCore(t, s)
}

func TestMatchExpressionWithDefaultNull(t *testing.T) {
	s := `val noMatch2 = match (999) {
		1 => { "one" }
	}
	assert(noMatch2 == null)`
	vmStringWithCore(t, s)
}

func TestVmErrorTryCatchScenario1(t *testing.T) {
	s := `fun riskyOperation(shouldFail) {
		try {
			if (shouldFail) {
				error("operation failed")
			}
			return "success"
		} catch (e) {
			return "failed: #{e}"
		}
	}

	println("riskyOperation = #{riskyOperation(true)}")
	assert(riskyOperation(true) == "failed: operation failed")`
	vmStringWithCore(t, s)
}

func TestInfiniteLoopInTryCatchScenario(t *testing.T) {
	s := `var outerCaught = false
	var innerCaught = false
	try {
		try {
			error("inner error")
		} catch (e) {
			println("caught inner")
			innerCaught = true
			assert(e == "inner error")
		}
	} catch (e) {
		println("caught outer")
		outerCaught = true
	}
	println('a')`
	vmStringWithCore(t, s)
}

func TestShadowingIssueWithListComprehension(t *testing.T) {
	s := `var x = 90;
	var __internal1__ = [];
	for (x in 1..5) {
		var __result__ = x ** 2;
		__internal1__ << __result__;
	}
	assert(__internal1__ == [1, 4, 9, 16, 25]);`
	vmStringWithCore(t, s)
}

func TestForInShadowsOuterVariable(t *testing.T) {
	s := `var x = "outer";
	fun test() {
		assert(x == "outer");
		var results = [];
		for (x in 1..3) {
			results << x;
		}
		assert(x == "outer");
		assert(results == [1, 2, 3]);
	}
	test();
	assert(x == "outer");`
	vmStringWithCore(t, s)
}

func TestOffByOneFloorDivAndModulo(t *testing.T) {
	s := `val c = -17
	val d = 5
	val result = (c // d) * d + (c % d);
	println(result);
	assert(c == (c // d) * d + (c % d))`
	vmStringWithCore(t, s)
}

func TestExpectedVmErrorForMapCompAddAndNoPanic(t *testing.T) {
	s := `val m1 = {a: 1, b: 2}
	val m2 = {c: 3, d: 4}
	val merged = {k: v for [k, v] in m1} + {k: v for [k, v] in m2}`
	vmStringWithCoreExpectErrorContaining(t, s, "unknown operator: MAP OpAdd MAP")
}

func TestAnotherVmScopeIssue(t *testing.T) {
	s := `var x = 1
	println('x1 = #{x}')
	if true {
		var x = 2
		println('x2 = #{x}')
		assert(x == 2)
		if true {
			var x = 3
			println('x3 = #{x}')
			assert(x == 3)
		}
		println('x2 = #{x}')
		assert(x == 2)
	}
	println('x1 = #{x}')
	assert(x == 1)`
	vmString(t, s)
}

func TestTrailingCommas(t *testing.T) {
	s := `fun hello(a,b,c,) {
		a + b + c
	}

	assert(hello(1,2,3) == 6);

	val x = [1,2,];
	assert(x == [1,2]);
	val y = {3,4,};
	assert(y == {3,4});
	val z = @{hello: 'world',};
	assert(z == @{hello: 'world'});
	println('hello = #{hello(1,2,3)}');
	println('z = #{z}');
	println(z.hello);
	val a = {aa: 'world',};
	assert(a == {aa: 'world'});`
	vmString(t, s)
}

func vmStringWithCore(t *testing.T, s string) {
	program := parseString(t, s)
	c := compiler.NewFromCore()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err.Error())
	}
	// fmt.Print(blueutil.BytecodeDebugString(c.Bytecode().Instructions, c.Bytecode().Constants))
	// fmt.Printf("PARSED: ```%s```\n", program.String())
	v := vm.New(c.Bytecode())
	err = v.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err.Error())
	}
	obj := v.LastPoppedStackElem()
	if obj.Type() == object.ERROR_OBJ {
		t.Fatalf("vm returned error: %s, %s", s, obj.(*object.Error).Message)
	}
}

func vmStringWithCoreExpectErrorContaining(t *testing.T, s, expectedErrorString string) {
	program := parseString(t, s)
	c := compiler.NewFromCore()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err.Error())
	}
	// fmt.Print(blueutil.BytecodeDebugString(c.Bytecode().Instructions, c.Bytecode().Constants))
	// fmt.Printf("PARSED: ```%s```\n", program.String())
	v := vm.New(c.Bytecode())
	err = v.Run()
	if err != nil {
		if !strings.Contains(err.Error(), expectedErrorString) {
			t.Fatalf("vm error did not contain expected: `%s`, got: `%s`", expectedErrorString, err.Error())
		}
	} else {
		t.Fatalf("vm expected error but got nothing")
	}
}

func vmString(t *testing.T, s string) {
	program := parseString(t, s)
	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err.Error())
	}
	// fmt.Print(cmd.BytecodeDebugString(c.Bytecode().Instructions, c.Bytecode().Constants))
	v := vm.New(c.Bytecode())
	err = v.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err.Error())
	}
	obj := v.LastPoppedStackElem()
	if obj.Type() == object.ERROR_OBJ {
		t.Fatalf("vm returned error: %s, %s", s, obj.(*object.Error).Message)
	}
}

func parseString(t *testing.T, s string) *ast.Program {
	l := lexer.New(s, "<string>")

	p := parser.New(l)
	program := p.ParseProgram()
	if p.HasErrors() {
		p.PrintParserErrors(os.Stderr)
		t.Fatalf("failed to parse string: %s", s)
	}
	return program
}

func compileStringWithCoreExpectCompilerErrorContaining(t *testing.T, s, expectedErrorString string) {
	t.Helper()
	program := parseString(t, s)
	c := compiler.NewFromCore()
	err := c.Compile(program)
	if err == nil {
		t.Fatalf("expected compiler error containing %q but got none", expectedErrorString)
	}
	if !strings.Contains(err.Error(), expectedErrorString) {
		t.Fatalf("compiler error %q does not contain expected %q", err.Error(), expectedErrorString)
	}
}

func compileStringWithCoreExpectSuccess(t *testing.T, s string) {
	t.Helper()
	program := parseString(t, s)
	c := compiler.NewFromCore()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("unexpected compiler error: %s", err.Error())
	}
}

func TestVarRedeclarationSameScope(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		errMatch string
	}{
		{"var then var", `var x = 0; var x = 1`, "already defined"},
		{"var then val", `var x = 0; val x = 1`, "already defined"},
		{"val then var", `val x = 0; var x = 1`, "already defined"},
		{"val then val", `val x = 0; val x = 1`, "already defined"},
		{"in nested block same level", `if true { var y = 1; var y = 2 }`, "already defined"},
		{"triple redeclaration", `var x = 0; var x = 1; var x = 2`, "already defined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compileStringWithCoreExpectCompilerErrorContaining(t, tt.input, tt.errMatch)
		})
	}
}

func TestVarShadowingDifferentScopes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"var then var in if", `var x = 0; if true { var x = 1 }`},
		{"var then val in if", `var x = 0; if true { val x = 1 }`},
		{"val then var in if", `val x = 0; if true { var x = 1 }`},
		{"val then val in if", `val x = 0; if true { val x = 1 }`},
		{"deeply nested", `var x = 0; if true { if true { var x = 1 } }`},
		{"for loop body", `for (var i = 0; i < 1; i += 1) { var i = 2 }`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compileStringWithCoreExpectSuccess(t, tt.input)
		})
	}
}

func TestVarBuiltinShadowing(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"shadow input", `val input = "test"`},
		{"shadow print", `var print = 42`},
		{"shadow type", `val type = "text"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compileStringWithCoreExpectSuccess(t, tt.input)
		})
	}
}

func TestVarShadowingAtRuntime(t *testing.T) {
	s := `var x = "outer"
	if true {
		var x = "inner"
		assert(x == "inner")
	}
	assert(x == "outer")

	var a = 1
	if true {
		var a = 2
		if true {
			var a = 3
			assert(a == 3)
		}
		assert(a == 2)
	}
	assert(a == 1)

	var mutable = "mutable"
	if true {
		val mutable = "immutable inside"
		assert(mutable == "immutable inside")
	}
	assert(mutable == "mutable")`
	vmStringWithCore(t, s)
}
