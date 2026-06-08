
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
