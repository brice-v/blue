# Comprehensive try-catch tests

# 1. Basic try-catch
var caught1 = false
try {
    error("basic error")
} catch (e) {
    caught1 = true
    assert(e == "basic error")
}
assert(caught1)

# 2. Try-catch-finally
var caught2 = false
var finallyRan2 = false
try {
    error("error in try")
} catch (e) {
    caught2 = true
    assert(e == "error in try")
} finally {
    finallyRan2 = true
}
assert(caught2)
assert(finallyRan2)

# 3. Try without error (catch skipped, finally runs)
var finallyRan3 = false
try {
    val x = 1 + 1
    assert(x == 2)
} catch (e) {
    assert(false)
} finally {
    finallyRan3 = true
}
assert(finallyRan3)

# 4. Try-finally only (no catch, no error)
var finallyRan4 = false
try {
    val a = 42
} finally {
    finallyRan4 = true
}
assert(finallyRan4)

# 5. Nested try-catch
var innerCaught = false
var outerCaught = false
try {
    try {
        error("inner error")
    } catch (e) {
        innerCaught = true
        assert(e == "inner error")
    }
} catch (e) {
    outerCaught = true
}
assert(innerCaught)
assert(!outerCaught)

# 6. Try-catch in function with return
fun testReturnInCatch() {
    try {
        error("fail")
    } catch (e) {
        return "caught: #{e}"
    }
    return "unreachable"
}
assert(testReturnInCatch() == "caught: fail")

# 7. Try-catch in function without error
fun testNoError() {
    try {
        return "success"
    } catch (e) {
        return "fail"
    }
}
assert(testNoError() == "success")

# 8. Multiple catches in sequence
var errors = []
for (i in 1..3) {
    try {
        error("error #{i}")
    } catch (e) {
        errors << e
    }
}
assert(len(errors) == 3)
assert(errors[0] == "error 1")
assert(errors[1] == "error 2")
assert(errors[2] == "error 3")

# 9. Continue inside catch in a loop
var count = 0
for (i in 0..5) {
    try {
        error("err")
    } catch (e) {
        count += 1
        continue
    }
    assert(false)
}
assert(count == 6)

# 10. Break inside catch
var breakCount = 0
for (i in 0..10) {
    try {
        error("err")
    } catch (e) {
        breakCount += 1
        break
    }
}
assert(breakCount == 1)

# 11. Try-catch with assert failure
var caughtAssert = false
try {
    assert(false, "assertion failed")
} catch (e) {
    caughtAssert = true
    assert(e.contains("assertion failed") || e.contains("assert"))
}
assert(caughtAssert)

# 12. Empty catch block
var caughtEmpty = false
try {
    error("silent")
} catch (e) {}
assert(true)

# 13. Empty finally block
try {
    val b = 1
} catch (e) {
    assert(false)
} finally {}
assert(true)

# 14. Nested try-catch with finally on both
var innerFinally15 = false
var outerFinally15 = false
try {
    try {
        error("nested")
    } catch (e) {
        assert(e == "nested")
    } finally {
        innerFinally15 = true
    }
} finally {
    outerFinally15 = true
}
assert(innerFinally15)
assert(outerFinally15)

# 15. Error in catch propagates to outer try
var outerCaught16 = false
try {
    try {
        error("first")
    } catch (e) {
        assert(e == "first")
        error("from catch")
    }
} catch (e) {
    outerCaught16 = true
    assert(e == "from catch")
}
assert(outerCaught16)

# 16. Catch variable is scoped to catch block
var catchVarScope = "outer"
try {
    error("test")
} catch (e) {
    catchVarScope = e
}
assert(catchVarScope == "test")

# 17. Try-catch with type error from invalid operation
var caughtType = false
try {
    val s = "hello" - 1
} catch (e) {
    caughtType = true
    assert(e.contains("type") || e.contains("subtraction"))
}
assert(caughtType)

# 18. Try-catch with division by zero
var caughtDiv = false
try {
    val z = 1 / 0
} catch (e) {
    caughtDiv = true
    assert(e.contains("zero") || e.contains("division"))
}
assert(caughtDiv)

assert(true)
