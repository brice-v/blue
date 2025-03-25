var x = @{one: 1, hello_world: "Hello World!"};

println(x);
assert(x == @{one: 1, hello_world: "Hello World!"});
x.one = 101;
x.hello_world = "abc";
println(x)
assert(x == @{one: 101, hello_world: "abc"});
# Test Immutability
val z = @{one: 123};
try {
    z.one = 99;
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "EvaluatorError: 'z' is immutable");
}
# Test Errors on Set
try {
    x.one = "Hello";
    assert(false, "UNREACHABLE");
} catch (e) {
    println(e);
    assert(e == "EvaluatorError: failed to set on struct literal: existing value type = INTEGER, new value type = STRING");
}
try {
    x.1 = 123;
    assert(false, "UNREACHABLE");
} catch (e) {
    println(e);
    assert(e == "EvaluatorError: index operator not supported: BLUE_STRUCT.1");
}
try {
    x.abc = null;
    assert(false, "UNREACHABLE");
} catch (e) {
    println(e);
    assert(e == "EvaluatorError: field name `abc` not found on blue struct: @{one: 101, hello_world: abc}");
}

assert(true);
