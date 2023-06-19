fun hello(arg1, arg2, arg3) {
    println(arg1, arg2, arg3)
}

try {
    hello();
    assert(false, "UNREACHABLE")
} catch (e) {
    assert("EvaluatorError: function called without enough arguments" in e);
}

fun hello1(arg1, arg2="bb", arg3) {
    println(arg1, arg2, arg3)
}

try {
    hello1("Hello");
    assert(false, "UNREACHABLE");
} catch (e) {
    assert("EvaluatorError: function called without enough arguments" in e);
}

fun hello2(arg1, arg2="aa", arg3, arg4) {
    println(arg1, arg2, arg3, arg4)
}

try {
    hello2("Hello", "SOME", "ANOTHER");
} catch (e) {
    assert(false, "UNREACHABLE");
}

assert(true);