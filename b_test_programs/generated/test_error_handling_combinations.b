# Test error handling combinations: try-catch-finally

# Basic try-catch
var caught = false
try {
    error("test error")
} catch (e) {
    caught = true
    assert(e == "test error")
}
assert(caught)

# Try-catch-finally
var finallyRan = false
try {
    error("error in try")
} catch (e) {
    assert(e == "error in try")
} finally {
    finallyRan = true
}
assert(finallyRan)

# Try without error, finally still runs
var finallyRan2 = false
try {
    val x = 1 + 1
    assert(x == 2)
} catch (e) {
    assert(false, "should not catch")
} finally {
    finallyRan2 = true
}
assert(finallyRan2)

### Not yet supported
# Nested try-catch
var outerCaught = false
var innerCaught = false
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

# Multiple catches
var caught1 = false
var caught2 = false
try {
    try {
        error("first error")
    } catch (e) {
        caught1 = true
        assert(e == "first error")
    }

    try {
        error("second error")
    } catch (e) {
        caught2 = true
        assert(e == "second error")
    }
} catch (e) {
    assert(false, "should not reach outer catch")
}
assert(caught1)
assert(caught2)

# Try with multiple finally blocks
var finally1 = false
var finally2 = false
try {
    try {
        error("nested error")
    } catch (e) {
        assert(e == "nested error")
    } finally {
        finally1 = true
    }
} catch (e) {
    assert(false, "should not reach")
} finally {
    finally2 = true
}
assert(finally1)
assert(finally2)
###

# Try-catch with return
fun testReturn() {
    try {
        error("error")
    } catch (e) {
        return "caught"
    } finally {
        # should still run
    }
    return "unreachable"println("GER")
}
assert(testReturn() == "caught")

# Try-catch in function
fun riskyOperation(shouldFail) {
    try {
        if (shouldFail) {
            error("operation failed")
        }
        return "success"
    } catch (e) {
        return "failed: #{e}"
    }
}

assert(riskyOperation(false) == "success")
# Not yet supported
#assert(riskyOperation(true) == "failed: operation failed")

# Try-catch with list operations
### This is a compiler error
var listError = false
try {
    val immutable = [1, 2, 3]
    immutable.push(4)
} catch (e) {
    listError = true
    assert("immutable" in e)
}
assert(listError)

# Try-catch with map operations
var mapError = false
try {
    val immutableMap = {a: 1}
    immutableMap["b"] = 2
} catch (e) {
    mapError = true
    assert("immutable" in e)
}
assert(mapError)
###

# Try-catch with division by zero
var divError = false
try {
    1 / 0
} catch (e) {
    divError = true
    assert(e.contains("division by zero") || e.contains("zero"))
}

assert(divError)

# Try-catch with invalid type
var typeError = false
try {
    val x = "hello"
    val y = x + 42
} catch (e) {
    typeError = true
    assert(e.contains("type") || e.contains("addition") || e.contains("string"))
}
assert(typeError)

# Try-catch-finally with variable scope
var finallyVar = null
try {
    var tryVar = "in try"
    error("test")
} catch (e) {
    finallyVar = "in catch"
} finally {
    # tryVar not accessible here
}
assert(finallyVar == "in catch")

# Try-catch with assertion
try {
    assert(false, "assertion failed")
} catch (e) {
    assert(e.contains("assertion failed"))
}

# Try-catch with multiple error types
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

# Assert with message
try {
    assert(1 + 1 == 3, "math is broken")
    assert(false, "should not reach")
} catch (e) {
    assert(e.contains("math is broken") || e.contains("assert"))
}
assert(true);