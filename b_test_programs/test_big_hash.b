#VM IGNORE
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
for (j in 1..10) {
    var z = "def";
}
try {
    z = 'zzz';
    assert(false, "UNREACHABLE");
} catch (e) {
    println("e = #{e}")
    assert("UNREACHABLE" notin e);
}


val xyz = 100.210389218302108380123 * 2.028140812;
println("xyz = #{xyz}");
val z = set([xyz]);
println("AFTER SET")
assert(true);

if (xyz in z) {
    assert(true);
} else {
    println("xyz = #{xyz}, z = #{z}");
    error("xyz not in z");
}


if (xyz notin z) {
    error("xyz should be in z");
} else {
    assert(true);
}

assert(true);