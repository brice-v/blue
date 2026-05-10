# Test nested scopes and variable shadowing

# Basic scope test
var outer = 100
if true {
    var inner = 200
    assert(inner == 200)
    assert(outer == 100)
}
assert(outer == 100)
### This is a compiler error
# inner should not be accessible here
try {
    println(inner)
    assert(false, "inner should not be accessible")
} catch (e) {
    assert(e.contains("identifier not found") || e.contains("not found"))
}
###

# Deep nesting
if true {
    var level1 = 1
    if true {
        var level2 = 2
        if true {
            var level3 = 3
            assert(level1 + level2 + level3 == 6)
        }
        assert(level1 + level2 == 3)
    }
    assert(level1 == 1)
}

### Currently Broken
# Shadowing at multiple levels
var x = 1
if true {
    var x = 2
    assert(x == 2)
    if true {
        var x = 3
        assert(x == 3)
        assert(x == 3)
    }
    assert(x == 2)
}
assert(x == 1)
###

# Function scope
fun testScope() {
    var local = 42
    return local
}
assert(testScope() == 42)

# Function with nested scope
fun testNested() {
    var a = 1
    if true {
        var b = 2
        var c = 3
        return a + b + c
    }
}
assert(testNested() == 6)

# Shadowing in function parameters
fun shadow(x) {
    var x = 100
    return x
}
assert(shadow(5) == 100)

# Multiple shadowing in same scope is an error
try {
    if true {
        var y = 1
        var y = 2
    }
    assert(false, "should have errored")
} catch (e) {
    assert("already defined" in e)
}

# Closure captures outer scope correctly
fun makeScopeTest() {
    var x = {counter: 0}
    return fun() {
        x.counter += 1
        return x.counter
    }
}

var counter1 = makeScopeTest()
var counter2 = makeScopeTest()
assert(counter1() == 1)
assert(counter2() == 1)
assert(counter1() == 2)
assert(counter2() == 2)

### Currently broken
# Shadowing in if/else branches
var branch = "outer"
if (true) {
    var branch = "inner"
    assert(branch == "inner")
}
assert(branch == "outer")
###

# Shadowing in for loop
for (var i = 0; i < 3; i += 1) {
    var i = 999  # This shadows the loop var inside the block
    assert(i == 999)
}
# After the loop, i should still be the loop variables final value
# Actually in blue, the for loop var is in the same scope as the body

# Multiple variables in one declaration
var [a,b,c] = [1,2,3]
assert(a == 1)
assert(b == 2)
assert(c == 3)

# Val cannot be reassigned in same scope
var mutable = 10
mutable = 20
assert(mutable == 20)

### This is a compiler error
# Immutable val cannot be reassigned
try {
    val immutable = 10
    immutable = 20
    assert(false, "should have errored")
} catch (e) {
    assert(e.contains("immutable") || e.contains("already defined"))
}
###

# Scope with try-catch
try {
    var tryVar = "in try"
    assert(tryVar == "in try")
    error("test error")
} catch (e) {
    ### This is a compiler error
    # tryVar should not be accessible here
    try {
        println(tryVar)
        assert(false, "tryVar should not be accessible")
    } catch (inner) {
        assert(inner.contains("identifier not found") || inner.contains("not found"))
    }
    ###
}

# Scope with match
val result = match (true) {
    true => {
        var matchVar = "in match"
        assert(matchVar == "in match")
        "matched"
    },
    _ => { "no match" },
}
assert(result == "matched")

# Variable accessible after for loop
for (var loopVar = 0; loopVar < 5; loopVar += 1) {
    # loop body
}
# loopVar is not supposed to be accessible outside the loop, this is a compiler error as expected
#assert(loopVar == 5)