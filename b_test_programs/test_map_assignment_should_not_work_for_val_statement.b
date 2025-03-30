val x = {hello: 'world'};

try {
    x.name = 'b';
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "EvaluatorError: 'x' is immutable");
}

val msg = "x should be {hello: 'world'}, got #{x}";
assert(x == {hello: 'world'}, msg);