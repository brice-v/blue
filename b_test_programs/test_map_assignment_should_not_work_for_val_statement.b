#VM IGNORE
# TODO: Implement checks for compiler
val x = {hello: 'world'};

try {
    x.name = 'b';
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "'x' is immutable");
}

val msg = "x should be {hello: 'world'}, got #{x}";
assert(x == {hello: 'world'}, msg);