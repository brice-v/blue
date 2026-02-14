#VM IGNORE
# TODO: vm will support destructuring
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

val xxx = {name: 'b', xyz: 123, 1: 'Some Var here'};
var yyy = {dd: 0x123, abcd: [1,2,3,5]}

val {name: myvar1, 'xyz': myvar2} = xxx;
var {dd: myvar3, abcd: myvar4} = yyy;

assert(myvar1 == 'b');
assert(myvar2 == 123);
assert(myvar3 == 0x123);
assert(myvar4 == [1,2,3,5])

var {name, xyz, '1': myVar} = {name: 'b', xyz: 123, '1': 'Some Var here'};
val {bbb, www, '2': myCustom} = {bbb: 'aaa', 'www': 0x123, '2': 'HERES ANOTHER'};

assert(name == 'b');
assert(xyz == 123);
assert(myVar == 'Some Var here');
assert(bbb == 'aaa');
assert(www == 0x123);
assert(myCustom == 'HERES ANOTHER')