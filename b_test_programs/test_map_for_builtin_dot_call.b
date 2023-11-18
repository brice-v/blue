var a = [1,2,3].type();
var b = {'a': 1, 'b': 2, 'c': 3}.type();
var c = {1,2,3}.type();

var d = [x for x in 1..10].type();
var e = {x: 'a' for x in 1..10}.type();
var f = {x for x in 1..10}.type();


println("a = #{a}");
println("b = #{b}");
println("c = #{c}");
println("d = #{d}");
println("e = #{e}");
println("f = #{f}");
assert(a == Type.LIST);
assert(b == Type.MAP);
assert(c == Type.SET);
assert(d == Type.LIST);
assert(e == Type.MAP);
assert(f == Type.SET);


[1,2,3].println();
{'a': 1, 'b': 2, 'c': 3}.println();
{1,2,3}.println();

[x for x in 1..10].println();
{x: 'a' for x in 1..10}.println();
{x for x in 1..10}.println();


assert(true);