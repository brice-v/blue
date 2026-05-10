# Test anonymous function invocation and inline functions

# Immediately invoked function expression
val result = fun() {
    42
}()
assert(result == 42)

# IIFE with parameters
val doubled = fun(x) {
    x * 2
}(21)
assert(doubled == 42)

# IIFE with multiple parameters
val sum = fun(a, b) {
    a + b
}(10, 32)
assert(sum == 42)

# IIFE with block body
val blockResult = fun() {
    var x = 1
    var y = 2
    var z = 3
    x + y + z
}()
assert(blockResult == 6)

# IIFE returning function
val getter = fun() {
    var x = {counter: 0}
    return fun() {
        x.counter += 1
        x.counter
    }
}()
assert(getter() == 1)
assert(getter() == 2)
assert(getter() == 3)

# Nested IIFE
val nested = fun() {
    fun() {
        100
    }()
}()
assert(nested == 100)

# IIFE in list
val list = [
    fun() { 1 }(),
    fun() { 2 }(),
    fun() { 3 }(),
]
assert(list == [1, 2, 3])

# IIFE in map
val map = {
    one: (fun() { 1 })(),
    two: (fun() { 2 })(),
    three: (fun() { 3 })(),
}
assert(map["one"] == 1)
assert(map["two"] == 2)
assert(map["three"] == 3)

# IIFE as function argument
fun apply(fn) {
    fn()
}

val applied = apply(fun() { "hello" })
assert(applied == "hello")

# IIFE with closure captured in outer scope
var captured = 0
fun makeCaptured() {
    var x = 10
    captured = fun() { x }()
}
makeCaptured()
assert(captured == 10)

# IIFE with default args
val withDefault = fun(x, y = 10) {
    x + y
}(5)
assert(withDefault == 15)

val withExplicit = fun(x, y = 10) {
    x + y
}(5, 20)
assert(withExplicit == 25)

# IIFE with condition
val conditional = fun() {
    if (true) {
        "yes"
    } else {
        "no"
    }
}()
assert(conditional == "yes")

# IIFE with try-catch
val safe = fun() {
    try {
        error("test")
    } catch (e) {
        return "caught";
    }
}()
assert(safe == "caught")

val safe1 = fun() {
    try {
        error("test")
    } catch (e) {
        "caught";
    }
}()
assert(safe1 == null);

# IIFE with for loop
var total = 0
fun() {
    for (i in 1..5) {
        total += i
    }
}()
assert(total == 15)

# IIFE with return
val earlyReturn = fun() {
    return 99
    return 100  # unreachable
}()
assert(earlyReturn == 99)

# IIFE in comprehension
val squares = [fun(x) { x * x }(i) for i in 1..5]
assert(squares == [1, 4, 9, 16, 25])

# IIFE with match
val matchResult = fun() {
    val m = match (5) {
        5 => { "five" },
        _ => { "other" },
    }
    return m
}()
assert(matchResult == "five")

# IIFE assigning to variable first
val myFunc = fun(x) { x + 1 }
val called = myFunc(10)
assert(called == 11)

# Chained IIFE
val chained = fun(x) {
    fun(y) {
        x + y
    }(20)
}(10)
assert(chained == 30)

# IIFE with self reference
val pidResult = fun() {
    val p = self()
    p.id
}()
assert(pidResult >= 0)

# IIFE with destructuring
val destruct = fun() {
    val [a, b, c] = [1, 2, 3]
    a + b + c
}()
assert(destruct == 6)

# IIFE with eval
val evalResult = fun() {
    eval("1 + 2 + 3")
}()
assert(evalResult == 6)

# IIFE with defer
var deferred = false
fun testDefer() {
    var p = fun() {
        defer(fun() { deferred = true })
    }
    p()
}
testDefer()
assert(deferred)