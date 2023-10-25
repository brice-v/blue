for var i = 0; i < 4; i += 1 {
    println("i = #{i}")
}

try {
    i = 100;
    assert(false, "unreachable");
} catch (e) {
    println("e = #{e}");
    assert("unreachable" notin e);
}

for (var i = 2; i < 10; i *= 2) {
    println("i = #{i}")
}

try {
    i = 100;
    assert(false, "unreachable");
} catch (e) {
    println("e = #{e}");
    assert("unreachable" notin e);
}

for (var i = 2; i < 10; i *= 2) {
    for (var j = 0; j < 4; j += 1) {
        println("i = #{i}, j = #{j}");
    }
    try {
        j = 100;
        assert(false, "unreachable");
    } catch (e) {
        println("e = #{e}");
        assert("unreachable" notin e);
    }
}
assert(true);