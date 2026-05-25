
var main = {
    some: fun(x, y, z=100) { x + y + z},
    other: fun(x=101, a, b=150) { x + a + b},
}

println(main.some(z=200, 1, 2))
assert(main.some(1,2, z=200) == 203);
assert(main.other(b=1, x=1, 3) == 5);