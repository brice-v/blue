val x = [1,2,3];

try {
    x.push!(4);
    assert(false, "UNREACHABLE")
} catch (e) {
    println(e);
    assert(e != "EvaluatorError: `assert` failed: UNREACHABLE");
    assert(e == "EvaluatorError: 'x' is immutable");
}

try {
    x.pop!();
    assert(false, "UNREACHABLE")
} catch (e) {
    println(e);
    assert(e != "EvaluatorError: `assert` failed: UNREACHABLE");
    assert(e == "EvaluatorError: 'x' is immutable");
}

try {
    x.shift!();
    assert(false, "UNREACHABLE")
} catch (e) {
    println(e);
    assert(e != "EvaluatorError: `assert` failed: UNREACHABLE");
    assert(e == "EvaluatorError: 'x' is immutable");
}

try {
    x.unshift!(4);
    assert(false, "UNREACHABLE")
} catch (e) {
    println(e);
    assert(e != "EvaluatorError: `assert` failed: UNREACHABLE");
    assert(e == "EvaluatorError: 'x' is immutable");
}

try {
    x[0] = 10;
    assert(false, "UNREACHABLE");
} catch (e) {
    println(e);
    assert(e != "EvaluatorError: `assert` failed: UNREACHABLE");
    assert(e == "EvaluatorError: 'x' is immutable");
}

val y = "Hello World";
try {
    y[0] = 'a';
    assert(false, "UNREACHABLE");
} catch (e) {
    println(e);
    assert(e != "EvaluatorError: `assert` failed: UNREACHABLE");
    assert(e == "EvaluatorError: 'y' is immutable");
}

var z = y;
z[0] = 'a';
assert(z == "aello World");


try {
    x << 4;
    assert(false, "UNREACHABLE");
} catch (e) {
    println(e);
    assert(e != "EvaluatorError: `assert` failed: UNREACHABLE");
    assert(e == "EvaluatorError: 'x' is immutable");
}

try {
    << x;
    assert(false, "UNREACHABLE");
} catch (e) {
    println(e);
    assert(e != "EvaluatorError: `assert` failed: UNREACHABLE");
    assert(e == "EvaluatorError: 'x' is immutable");
}