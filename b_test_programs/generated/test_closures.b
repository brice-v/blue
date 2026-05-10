# Test closures - functions capturing outer scope variables

# Basic closure
fun makeAdder(n) {
    return fun(x) {
        x + n
    }
}

val add10 = makeAdder(10)
val add5 = makeAdder(5)

assert(add10(3) == 13)
assert(add5(7) == 12)
assert(add10(0) == 10)

# Closure with multiple captured variables
fun makeMultiplier(a, b) {
    return fun(x) {
        x * a * b
    }
}

val mult6 = makeMultiplier(2, 3)
assert(mult6(5) == 30)

# Closure with mutable captured variable
fun makeCounter() {
    var x = {count: 0}
    return fun() {
        x.count += 1
        x.count
    }
}

val counter = makeCounter()
assert(counter() == 1)
assert(counter() == 2)
assert(counter() == 3)

# Nested closures - closure returning a closure
fun outer(x) {
    var y = 10
    return fun(a) {
        var z = 100
        return fun(b) {
            x + y + a + b + z
        }
    }
}

val inner = outer(1)
val deepest = inner(2)
assert(deepest(3) == 116)  # 1 + 10 + 2 + 3 + 100

# Closure in list
fun makeOps() {
    val ops = []
    ops << fun(x) { x + 1 }
    ops << fun(x) { x * 2 }
    ops << fun(x) { x ** 2 }
    return ops
}

val ops = makeOps()
assert(ops[0](5) == 6)
assert(ops[1](5) == 10)
assert(ops[2](5) == 25)

# Closure with list mutation
fun makeAppender() {
    var items = []
    return fun(item) {
        items << item
        len(items)
    }
}

val appender = makeAppender()
assert(appender("a") == 1)
assert(appender("b") == 2)
assert(appender("c") == 3)

# Closure as map value
fun makeMapWithClosures() {
    var base = 100
    return {
        add: fun(x) { x + base },
        mul: fun(x) { x * base },
    }
}

val m = makeMapWithClosures()
assert(m["add"](5) == 105)
assert(m["mul"](5) == 500)

# Closure with default args
fun makeDefaultAdder(n) {
    return fun(x, y = n) {
        x + y
    }
}

val adder = makeDefaultAdder(10)
assert(adder(5) == 15)
assert(adder(5, 20) == 25)

# Closure capturing from for loop
fun makeClosures() {
    val closures = []
    for (i in 1..5) {
        closures << fun(x) { x + i }
    }
    return closures
}

val cl = makeClosures()
assert(cl[0](0) == 1)
assert(cl[4](0) == 5)

# Closure passed as argument
fun applyTwice(fn, v) {
    fn(fn(v))
}

val double = fun(x) { x * 2 }
assert(applyTwice(double, 3) == 12)  # 3 -> 6 -> 12

### Not yet supported
# Closure with try-catch
fun makeSafeAdder() {
    var safe = true
    return fun(a, b) {
        try {
            a + b
        } catch (e) {
            if (safe) {
                0
            } else {
                error(e)
            }
        }
    }
}

val safeAdder = makeSafeAdder()
assert(safeAdder(1, 2) == 3)
###

# Closure returning closure returning closure
fun level1(a) {
    fun level2(b) {
        fun level3(c) {
            a + b + c
        }
        return level3
    }
    return level2
}

assert(level1(1)(2)(3) == 6)
