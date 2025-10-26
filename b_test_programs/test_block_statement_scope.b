var x = 123;

if (x == 123) {
    var y = "abc";
    if (y == "abc") {
        var abc = 999;
        println("abc should be found here #{abc}");
        assert(abc == 999);
        assert(x == 123);
    }
    println(y);
    assert(y == 'abc');
    assert(x == 123);
    try {
        assert(x == 123);
        println(abc);
    } catch (e) {
        assert(e == 'identifier not found: abc');
    }
}

try {
    println("y = #{y}");
    error("y should not be found");
} catch (e) {
    if ("y should not be found" in e) {
        error(e);
    }
}

for (i in 1..10) {
    var z = "def";
}

try {
    println("z = #{z}");
    error("z should not be found");
} catch (e) {
    if ("z should not be found" in e) {
        error(e);
    }
}

#println("x = #{x}, y = #{y}, z = #{z}");

true;