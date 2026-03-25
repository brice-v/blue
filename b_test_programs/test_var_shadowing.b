#VM IGNORE
var x = 1234;
try {
    val x = "HELLO";
    println(x);
    assert(false) # unreachable
} catch (e) {
    println(e);
    assert(e == "'x' is already defined");
}

val y = 1234;

try {
    var y = "HELLO";
    println(y);
    assert(false); # unreachable
} catch (e) {
    println(e);
    assert(e == "'y' is already defined as immutable, cannot reassign")
}

assert(true);