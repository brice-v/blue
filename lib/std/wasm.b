## `wasm` provides an init function which sets up a wasm module
## to be usable as an external process with side effects (ie. mod.run())
## or call functions within it (ie. mod.add(1,2) -> 3)
##
## Note: all wasm modules need to be compiled with wasi builtin so its
## properly initialized

val __wasm_init = _wasm_init;
val __wasm_run = _wasm_run;
val __wasm_get_functions = _wasm_get_functions;
val __wasm_get_exported_function = _wasm_get_exported_function;
val __wasm_close = _wasm_close;

fun init(wasm_code_path, args=ARGV, mounts={'.':'/'}, stdout=FSTDOUT, stderr=FSTDERR, stdin=FSTDIN, envs=ENV, enable_rand=true, enable_time_and_sleep_precision=true, host_logging='', listens=[], timeout=0) {
    ##std:this,__wasm_init,__wasm_run,__wasm_get_functions,__wasm_get_exported_function,__wasm_close
    ## `init` will take a wasm module compiled for wasi and allow it to run or export functions with the given context in the function parameters
    ##
    ## wasm_code_path: is the path to the .wasm file
    ## args: is a list of args to be passed to the wasm module, by default its the current ARGV of the current process
    ## mounts: is the available filesystem directories to the module, by default a map of the current directory '.' is mapped to the root directory of the module '/'
    ## stdout: is a *os.File for stdout writing within the module, by default its FSTDOUT
    ## stderr: is a *os.File for stderr writing within the module, by default its FSTDERR
    ## stdin: is a *os.File for stdin reading within the module, by default its FSTDIN
    ## envs: is a map of environment variables accessible to the module, by default its set to ENV
    ## enable_rand: is a bool which allows the module to use random, by default its set to true
    ## enable_time_and_sleep_precision: is a bool allowing all the time and sleep precision within the module, by default its set to true
    ## host_logging: is a string comma separated which will log module info, by default its empty (available options: all,clock,filesystem,memory,proc,poll,random,sock)
    ## listens: is a list of strings in the format 'addr:port' for the config of the module
    ## timeout: is an int as nanosecond count for how long the module can run for, by default its set to 0 which means no timeout
    ##
    ## see b_test_programs/test_wasm.b for some basic examples on how this function and related are used in context
    ##
    ## init(wasm_code_path, args=ARGV, mounts={'.':'/'}, stdout=FSTDOUT, stderr=FSTDERR, stdin=FSTDIN, envs=ENV, enable_rand=true, enable_time_and_sleep_precision=true, host_logging='', listens=[], timeout=0)
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