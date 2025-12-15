var s = "abc"
var expected = "abcabcabc"

assert(s * 3 == expected)
assert(3 * s == expected)
assert(0b11 * s == expected)
assert(s * 0b11 == expected)
println(s * 3)

var thislist = [0,1,2,3,4] + [0,1,2,3,4]
expected = [0,1,2,3,4,0,1,2,3,4]
assert(thislist == expected);