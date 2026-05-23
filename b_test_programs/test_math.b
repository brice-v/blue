import math

var r = 100 ** 0.5;
println(r);
println(math.sqrt(100));
if (r != 10.0) {
    assert(false)
}

assert(-7 // 2 == -4)
assert(7 // -2 == -4)
assert(-7 // -2 == 3)
# Modulo with negatives
assert(-7 % 2 == 1)
assert(7 % -2 == -1)
assert(-7 % -2 == -1)