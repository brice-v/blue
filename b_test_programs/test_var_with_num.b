
try {
    var abc123 = 1;
    val abcd123 = 2;
    println(abc123);
    println(abcd123);
} catch (e) {
    # Shouldnt reach this
    assert(false);
}

assert(true);