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
    assert(e == "EvaluatorError: failed to set on struct literal: existing value type = INTEGER, new value type = STRING");
}
try {
    x.1 = 123;
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "EvaluatorError: index operator not supported: BLUE_STRUCT.INTEGER");
}
try {
    x.abc = null;
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "EvaluatorError: field name `abc` does not exist on blue struct");
}

assert(true);
