var x = 123;

if (x == 123) {
    var y = "abc";
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