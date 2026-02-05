package b_program_test

import (
	"blue/ast"
	"blue/compiler"
	"blue/evaluator"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/repl"
	"blue/vm"
	"bytes"
	"fmt"
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

const vmScopes = `
if true {
	var a = 123;
}
a = 555;
assert(a != 555);
`

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

// TODO: Enable later on once more is done
func testAllProgramsInDirectoryWithVm(t *testing.T) {
	files, err := os.ReadDir("./")
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		executeBlueTestFileWithVm(f, t)
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

func executeBlueTestFileWithVm(f fs.DirEntry, t *testing.T) {
	if !strings.HasSuffix(f.Name(), ".b") {
		return
	}
	fpath, err := filepath.Abs(f.Name())
	if err != nil {
		t.Fatal(err)
	}
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
	defer openFile.Close()

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
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(os.Stderr, p.Errors())
		t.Fatalf("File `%s`: failed to parse", f.Name())
	}
	globals := make([]object.Object, vm.GlobalsSize)
	c := compiler.NewFromCore()
	err = c.Compile(program)
	if err != nil {
		t.Errorf("File `%s`: compiler returned error %s", f.Name(), err.Error())
		return
	}
	v := vm.NewWithGlobalsStore(c.Bytecode(), c.Tokens, globals)
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
	if obj.Inspect() != "true" {
		t.Errorf("File `%s`: Did not return true as last statement. Failed", f.Name())
	}
}

func TestFibPerf(t *testing.T) {
	execString(t, fibEx)
}

func execString(t *testing.T, s string) {
	program := parseString(t, s)
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

func testVmScopes(t *testing.T) {
	vmString(t, vmScopes)
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

func vmStringWithCore(t *testing.T, s string) {
	program := parseString(t, s)
	c := compiler.NewFromCore()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err.Error())
	}
	// fmt.Print(utils.BytecodeDebugString(c.Bytecode().Instructions, c.Bytecode().Constants))
	fmt.Printf("PARSED: ```%s```\n", program.String())
	v := vm.New(c.Bytecode(), c.Tokens)
	err = v.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err.Error())
	}
	obj := v.LastPoppedStackElem()
	if obj.Type() == object.ERROR_OBJ {
		t.Fatalf("vm returned error: %s, %s", s, obj.(*object.Error).Message)
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
	v := vm.New(c.Bytecode(), c.Tokens)
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
	if len(p.Errors()) != 0 {
		repl.PrintParserErrors(os.Stderr, p.Errors())
		t.Fatalf("failed to parse string: %s", s)
	}
	return program
}
