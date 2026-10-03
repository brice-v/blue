
fun o() {
    var this = {};
    this._x = 0;
    this._y = [];
    this.__type = || => "O object"
    this.incr = fun() {
        this._x += 1;
    }
    this.__str = fun() {
        "{x: #{this._x}}";
    }
    this.__add = fun(other) {
        this._x + other._x;
    }
    this.__sub = fun(other) {
        this._x - other._x;
    }
    this.__mul = fun(other) {
        this._x * other._x;
    }
    this.__div = fun(other) {
        this._x / other._x;
    }
    this.__mod = fun(other) {
        this._x % other._x;
    }
    this.__fdiv = fun(other) {
        this._x // other._x;
    }
    this.__pow = fun(other) {
        this._x ** other._x;
    }
    this.__and = fun(other) {
        "#{this} #{other} __and"
    }
    this.__or = fun(other) {
        "#{this} #{other} __or"
    }
    this.__xor = fun(other) {
        "#{this} #{other} __xor"
    }
    this.__rshift = fun(other) {
        "#{this} #{other} __rshift"
    }
    this.__lshift = fun(other) {
        this._y << other;
        "#{this} #{other} __lshift"
    }
    this.__neg = fun() {
        return -this._x;
    }
    this.__inv = fun() {
        return ~this._x;
    }
    this.__eq = fun(other) {
        return this._x == other._x;
    }
    this.__ne = fun(other) {
        return this._x != other._x;
    }
    this.__gt = fun(other) {
        return this._x > other._x;
    }
    this.__gte = fun(other) {
        return this._x >= other._x;
    }
    this.__len = fun() {
        return len(this._y);
    }
    return this;
}

val o1 = o();
o1.incr();
o1.incr();
o1.incr();
o1.incr();
o1.incr();
val abc1 = str(o1);
println(abc1);
val abc = "#{o1}";
println(abc);
println(o1);
val expected = "{x: 5}"
assert(abc == expected);
assert(abc1 == expected);
val o2 = o();
o2.incr();
assert(o1 + o2 == 6);
println("o1 - o2 = #{o1 - o2}");
assert(o1 - o2 == 4)
println("o1 * o2 = #{o1 * o2}");
assert(o1 * o2 == 5)
println("o1 / o2 = #{o1 / o2}")
assert(o1 / o2 == 5)
println("o1 % o2 = #{o1 % o2}")
assert(o1 % o2 == 0)
println("o1 // o2 = #{o1 // o2}")
assert(o1 // o2 == 5)
println("o1 ** o2 = #{o1 ** o2}")
assert(o1 ** o2 == 5)

println("---------")
println(o1 & o2)
assert((o1 & o2) == "{x: 5} {x: 1} __and")
println(o1 | o2)
assert((o1 | o2) == "{x: 5} {x: 1} __or")
println(o1 ^ o2)
assert((o1 ^ o2) == "{x: 5} {x: 1} __xor")
println(o1 >> o2)
assert((o1 >> o2) == "{x: 5} {x: 1} __rshift")
println(o1 << o2)
assert((o1 << o2) == "{x: 5} {x: 1} __lshift")


println("o1 = #{o1}");
println("o2 = #{o2}");
println("o1 == o2 = #{o1 == o2}");
assert((o1 == o2) == false)
println("o1 != o2 = #{o1 != o2}");
assert((o1 != o2))
println("o1 > o2 = #{o1 > o2}");
assert((o1 > o2))
println("o1 >= o2 = #{o1 >= o2}");
assert((o1 >= o2))
println("o1 < o2 = #{o1 < o2}");
assert((o1 < o2) == false)
println("o1 <= o2 = #{o1 <= o2}");
assert((o1 <= o2) == false)

println("-o1 = #{-o1}");
assert(-o1 == -5)
println("~o1 = #{~o1}");
assert(~o1 == -6)

var a1 = o(); a1.incr(); a1.incr();
var a2 = o(); a2.incr();
a1 += a2;
assert(a1 == 3)

var s1 = o(); s1.incr(); s1.incr();
var s2 = o(); s2.incr();
s1 -= s2;
assert(s1 == 1)

var m1 = o(); m1.incr(); m1.incr();
var m2 = o(); m2.incr();
m1 *= m2;
assert(m1 == 2)

var d1 = o(); d1.incr(); d1.incr();
var d2 = o(); d2.incr();
d1 /= d2;
assert(d1 == 2)

var p1 = o(); p1.incr(); p1.incr();
var p2 = o(); p2.incr();
p1 %= p2;
assert(p1 == 0)

var f1 = o(); f1.incr(); f1.incr();
var f2 = o(); f2.incr();
f1 //= f2;
assert(f1 == 2)

var e1 = o(); e1.incr(); e1.incr();
var e2 = o(); e2.incr();
e1 **= e2;
assert(e1 == 2)

var an1 = o(); an1.incr(); an1.incr();
var an2 = o(); an2.incr();
an1 &= an2;
assert(an1 == "{x: 2} {x: 1} __and")

var or1 = o(); or1.incr(); or1.incr();
var or2 = o(); or2.incr();
or1 |= or2;
assert(or1 == "{x: 2} {x: 1} __or")

var x1 = o(); x1.incr(); x1.incr();
var x2 = o(); x2.incr();
x1 ^= x2;
assert(x1 == "{x: 2} {x: 1} __xor")

var r1 = o(); r1.incr(); r1.incr();
var r2 = o(); r2.incr();
r1 >>= r2;
assert(r1 == "{x: 2} {x: 1} __rshift")

var l1 = o(); l1.incr(); l1.incr();
var l2 = o(); l2.incr();
l1 <<= l2;
assert(l1 == "{x: 2} {x: 1} __lshift")

var ooo = o();
ooo << o();
ooo << o();
ooo << o();
assert(len(ooo) == 3);

assert(type(ooo) == "O object")

fun box() {
    var this = {};
    this._x = 10;
    this.__str = fun() {
        "box(#{this._x})";
    }
    this.__add = fun(other) {
        "add:#{this._x},#{other}";
    }
    this.__radd = fun(other) {
        "radd:#{this._x},#{other}";
    }
    this.__sub = fun(other) {
        "sub:#{this._x},#{other}";
    }
    this.__rsub = fun(other) {
        "rsub:#{this._x},#{other}";
    }
    this.__mul = fun(other) {
        "mul:#{this._x},#{other}";
    }
    this.__rmul = fun(other) {
        "rmul:#{this._x},#{other}";
    }
    this.__div = fun(other) {
        "div:#{this._x},#{other}";
    }
    this.__rdiv = fun(other) {
        "rdiv:#{this._x},#{other}";
    }
    this.__mod = fun(other) {
        "mod:#{this._x},#{other}";
    }
    this.__rmod = fun(other) {
        "rmod:#{this._x},#{other}";
    }
    this.__fdiv = fun(other) {
        "fdiv:#{this._x},#{other}";
    }
    this.__rfdiv = fun(other) {
        "rfdiv:#{this._x},#{other}";
    }
    this.__pow = fun(other) {
        "pow:#{this._x},#{other}";
    }
    this.__rpow = fun(other) {
        "rpow:#{this._x},#{other}";
    }
    this.__and = fun(other) {
        "and:#{this._x},#{other}";
    }
    this.__rand = fun(other) {
        "rand:#{this._x},#{other}";
    }
    this.__or = fun(other) {
        "or:#{this._x},#{other}";
    }
    this.__ror = fun(other) {
        "ror:#{this._x},#{other}";
    }
    this.__xor = fun(other) {
        "xor:#{this._x},#{other}";
    }
    this.__rxor = fun(other) {
        "rxor:#{this._x},#{other}";
    }
    this.__rshift = fun(other) {
        "rshift:#{this._x},#{other}";
    }
    this.__rrshift = fun(other) {
        "rrshift:#{this._x},#{other}";
    }
    this.__lshift = fun(other) {
        "lshift:#{this._x},#{other}";
    }
    this.__rlshift = fun(other) {
        "rlshift:#{this._x},#{other}";
    }
    this.__matmul = fun(other) {
        "matmul:#{this._x},#{other}";
    }
    this.__rmatmul = fun(other) {
        "rmatmul:#{this._x},#{other}";
    }
    this.__gt = fun(other) {
        "gt:#{this._x},#{other}";
    }
    this.__gte = fun(other) {
        "gte:#{this._x},#{other}";
    }
    this.__rgt = fun(other) {
        "rgt:#{this._x},#{other}";
    }
    this.__rgte = fun(other) {
        "rgte:#{this._x},#{other}";
    }
    return this;
}

fun checkDunder(actual, expected, label, fails) {
    if (actual != expected) {
        fails << "#{label}: expected #{expected}, got #{actual}";
    }
}

val b1 = box();
val b2 = box();

assert((b1 + 1) == "add:10,1");
assert((b1 - 1) == "sub:10,1");
assert((b1 * 2) == "mul:10,2");
assert((b1 / 2) == "div:10,2");
assert((b1 % 3) == "mod:10,3");
assert((b1 // 3) == "fdiv:10,3");
assert((b1 ** 2) == "pow:10,2");
assert((b1 & 3) == "and:10,3");
assert((b1 | 3) == "or:10,3");
assert((b1 ^ 3) == "xor:10,3");
assert((b1 >> 2) == "rshift:10,2");
assert((b1 << 2) == "lshift:10,2");
assert((b1 @ 2) == "matmul:10,2");
assert((b1 @ b2) == "matmul:10,box(10)");
assert((b1 > 5) == "gt:10,5");
assert((b1 >= 5) == "gte:10,5");
assert((5 < b1) == "gt:10,5");
assert((5 <= b1) == "gte:10,5");

val dunderFails = [];
checkDunder(1 + b1, "radd:10,1", "1 + b1", dunderFails);
checkDunder(1 - b1, "rsub:10,1", "1 - b1", dunderFails);
checkDunder(2 * b1, "rmul:10,2", "2 * b1", dunderFails);
checkDunder(2 / b1, "rdiv:10,2", "2 / b1", dunderFails);
checkDunder(3 % b1, "rmod:10,3", "3 % b1", dunderFails);
checkDunder(3 // b1, "rfdiv:10,3", "3 // b1", dunderFails);
checkDunder(2 ** b1, "rpow:10,2", "2 ** b1", dunderFails);
checkDunder(3 & b1, "rand:10,3", "3 & b1", dunderFails);
checkDunder(3 | b1, "ror:10,3", "3 | b1", dunderFails);
checkDunder(3 ^ b1, "rxor:10,3", "3 ^ b1", dunderFails);
checkDunder(2 >> b1, "rrshift:10,2", "2 >> b1", dunderFails);
checkDunder(2 << b1, "rlshift:10,2", "2 << b1", dunderFails);
checkDunder(2 @ b1, "rmatmul:10,2", "2 @ b1", dunderFails);
checkDunder(b1 < 5, "rgt:10,5", "b1 < 5", dunderFails);
checkDunder(b1 <= 5, "rgte:10,5", "b1 <= 5", dunderFails);
checkDunder(5 > b1, "rgt:10,5", "5 > b1", dunderFails);
checkDunder(5 >= b1, "rgte:10,5", "5 >= b1", dunderFails);
assert(len(dunderFails) == 0, str(dunderFails));