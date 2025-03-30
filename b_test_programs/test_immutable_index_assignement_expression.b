val x = {'hello': {a: {c: 1}}};


try {
    x.hello.a.c += 1
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "EvaluatorError: 'x' is immutable");
}

try {
    x.hello.a.c = 3;
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "EvaluatorError: 'x' is immutable");
}


val y = @{'hello': {a: @{c: 1}}};

try {
    y.hello.a.c += 1
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "EvaluatorError: 'y' is immutable");
}

try {
    y.hello.a.c = 3;
    assert(false, "UNREACHABLE");
} catch (e) {
    assert(e == "EvaluatorError: 'y' is immutable");
}

assert(true);