println("cwd = #{cwd()}")
import wasm

var prefix = "";
if ("b_test_programs" notin cwd()) {
    prefix = "b_test_programs/";
}

var mod = wasm.init(prefix+'wasm_test_files/cat.wasm', args=[prefix+'wasm_test_files/cat.go.tmp']);
var rc = mod.run();
assert(rc == 0);

var mod1 = wasm.init(prefix+'wasm_test_files/add.wasm');
defer(fun() { 
    try {
        mod1.close();
    } catch (e) {
        println("error closing mod1 #{e}");
    }
}); 
var functions = mod1.get_functions();
println("functions = #{functions}")
var expected_functions = ['realloc', '_start', 'add', 'asyncify_start_unwind', 'asyncify_stop_unwind', 'asyncify_start_rewind', 'free', 'calloc', 'asyncify_stop_rewind', 'malloc', 'asyncify_get_state'];
assert(len(functions) == len(expected_functions));
for (func in expected_functions) {
    assert(func in functions);
}
println(mod1.add(0x3, 0x7));
var result = mod1.add(0x3,0x7);
println(result[0]);
assert(result[0] == uint(10))


assert(true);