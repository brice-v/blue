
val x = {hello: 'world', another: 'ignored'};
val y = {d: 1, abc: [1,2,3]}

val { hello } = x;
assert(hello == 'world');

var { d, abc } = y;
assert(d == 1);
assert(abc == [1,2,3]);


val z = [1,2,3,4];

var [a,b] = z;
val [c, g, e, f] = z;

assert(a == 1);
assert(b == 2);
assert(c == 1);
assert(g == 2);
assert(e == 3);
assert(f == 4);