for (i in 0..10) {
    try {
        assert(false)
    } catch (e) {
        continue
    }
    assert(false);
}
assert(true);

var y = 0;
for (i in 0..10) {
    println("in here #{i}");
    try {
        assert(false)
    } catch (e) {}
    y += i;
}
println("y = #{y}");
assert(y == 55);

for (i in 0..10) {
    try {
        assert(false)
    } catch (e) {}
    finally {}
    y += i;
}

println("y = #{y}");
assert(y == 110);
