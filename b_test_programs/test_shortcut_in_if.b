var x = null;
var y = fun() { assert(false); };

if (x != null && y()) {
    println("SHOULD NOT PRINT")
    assert(false);
}

if (x == null || y()) {
    println("SHOULD PRINT")
    assert(true);
}
assert(true);