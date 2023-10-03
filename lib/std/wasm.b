## TODO: module description

val __wasm_init = _wasm_init;
val __wasm_run = _wasm_run;
val __wasm_get_functions = _wasm_get_functions;
val __wasm_get_exported_function = _wasm_get_exported_function;

fun init(wasm_code_path, args=ARGV, mounts={'.':'/'}, stdout=FSTDOUT, stderr=FSTDERR, stdin=FSTDIN, envs=ENV, enable_rand=true, enable_time_and_sleep_precision=true, host_logging='', listens=[], timeout=0) {
    ## TODO: Include docs
    var this = {};
    this.mod = __wasm_init(wasm_code_path, args, mounts, stdout, stderr, stdin, envs, enable_rand, enable_time_and_sleep_precision, host_logging, listens, timeout);
    this.run = fun() {
        __wasm_run(this.mod)
    };
    this.get_functions = fun() {
        __wasm_get_functions(this.mod)
    };
    val functions = __wasm_get_functions(this.mod);
    for (func in functions) {
        this[func] = __wasm_get_exported_function(this.mod, func);
    }
    this.close = fun() {
        __wasm_close(this.mod)
    };
    # Essentially if we can put the functions as a key here to use as a function thats the ideal scenario (and they could sort of be builtin functions)
    return this;
}