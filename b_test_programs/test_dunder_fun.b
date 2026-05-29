
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