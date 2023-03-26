for (i in 0..10) {
    try {
        x = 1;
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
        x = 1;
    } catch (e) {}
    y += i;
}
println("y = #{y}");
assert(y == 55);

for (i in 0..10) {
    try {
        x = 1;
    } catch (e) {}
    finally {}
    y += i;
}

println("y = #{y}");
assert(y == 110);
