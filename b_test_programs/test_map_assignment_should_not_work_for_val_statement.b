val x = {hello: 'world'};

try {
    x.name = 'b';
} catch (e) {
    assert(e == "EvaluatorError: 'x' is immutable");
}

val msg = "x should be {hello: 'world'}, got #{x}";
assert(x == {hello: 'world'}, msg);