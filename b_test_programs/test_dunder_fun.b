
fun o() {
    var this = {};
    this._x = 0;
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
