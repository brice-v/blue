fun hello(abc, thing=1, other=[], last) {
    println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
    "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
}

var x = {
    hello2: fun(abc, thing=1, other=[], last) {
        println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
        "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
    },
    gg: {
        hello3: fun(abc, thing=1, other=[], last) {
            println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
            "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
        },
    },
};
x.hello3 = fun(abc, thing=1, other=[], last) {
    println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
    "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
};
x.hello4 = hello;

var yy = [fun(abc, thing=1, other=[], last) {
    println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
    "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
}, [
    fun(abc, thing=1, other=[], last) {
        println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
        "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
    }
   ]
];
var zz = set([fun(abc, thing=1, other=[], last) {
    println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
    "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
}]);

yy << fun(abc, thing=1, other=[], last) {
    println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
    "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
};
zz << fun(abc, thing=1, other=[], last) {
    println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
    "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
};


val y = fun(abc, thing=1, other=[], last) {
    println("abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}")
    "abc = #{abc}, thing = #{thing}, other = #{other}, last = #{last}"
};


assert(hello(1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(hello(1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(hello(1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(hello("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(hello("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")
assert("ONE".hello("LAST") == "abc = ONE, thing = 1, other = [], last = LAST");

assert(x.hello3(1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(x.hello3(1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(x.hello3(1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(x.hello3("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(x.hello3("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")

assert(x.hello2(1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(x.hello2(1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(x.hello2(1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(x.hello2("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(x.hello2("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")

assert(x.hello4(1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(x.hello4(1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(x.hello4(1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(x.hello4("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(x.hello4("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")

assert(x.gg.hello3(1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(x.gg.hello3(1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(x.gg.hello3(1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(x.gg.hello3("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(x.gg.hello3("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")


assert(yy[0](1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(yy[0](1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(yy[0](1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(yy[0]("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(yy[0]("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")

assert(yy[1][0](1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(yy[1][0](1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(yy[1][0](1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(yy[1][0]("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(yy[1][0]("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")

assert(zz[0](1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(zz[0](1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(zz[0](1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(zz[0]("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(zz[0]("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")

assert(yy[2](1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(yy[2](1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(yy[2](1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(yy[2]("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(yy[2]("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")

assert(zz[1](1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(zz[1](1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(zz[1](1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(zz[1]("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(zz[1]("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")

assert(y(1, "LAST") == "abc = 1, thing = 1, other = [], last = LAST")
assert(y(1, thing='x', other=2, "LAST") == "abc = 1, thing = x, other = 2, last = LAST")
assert(y(1, other=1, thing=2, "LAST") == "abc = 1, thing = 2, other = 1, last = LAST")
assert(y("HELLO", thing=3, "LAST") == "abc = HELLO, thing = 3, other = [], last = LAST")
assert(y("HELLO", 2, null, "DONE") == "abc = HELLO, thing = 2, other = null, last = DONE")
assert("ONE".y("LAST") == "abc = ONE, thing = 1, other = [], last = LAST");