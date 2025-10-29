# VM IGNORE
for (i in 1..2) {
    var z = "def";
}
try {
    z = 'aaa';
    assert(false, "UNREACHABLE");
} catch (e) {
    println("e = #{e}")
    assert("UNREACHABLE" notin e);
}

var abca = [1, 2, 3];
var abcb = [];
for ([a, b] in abca) {
    println("a=#{a}, b=#{b}");
    abcb[a] = b;
}
println("DONE");
assert(true);