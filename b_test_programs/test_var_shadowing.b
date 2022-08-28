var x = 1234;
try {
    val x = "HELLO";
    assert(false) # unreachable
} catch (e) {
    assert(e == "EvaluatorError: 'x' is already defined");
}

val y = 1234;

try {
    var y = "HELLO";
    assert(false); # unreachable
} catch (e) {
    assert(e == "EvaluatorError: 'y' is already defined as immutable, cannot reassign")
}

return true;