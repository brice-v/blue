var s = "abc"
var expected = "abcabcabc"

if (s * 3 != expected) {
    false
}

if (3 * s != expected) {
    false
}

if (0b11 * s != expected) {
    false
}

if (s * 0b11 != expected) {
    false
}

println(s * 3)


var thislist = [0,1,2,3,4] + [0,1,2,3,4]
expected = [0,1,2,3,4,0,1,2,3,4]
if (thislist != expected) {
    false
}
true