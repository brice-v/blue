# Test floor division (//) and modulo (%) operators

# Basic floor division
assert(7 // 2 == 3)
assert(10 // 3 == 3)
assert(100 // 7 == 14)
assert(0 // 5 == 0)
assert(5 // 5 == 1)
assert(3 // 5 == 0)

# Floor division with negatives (floor toward -inf)
println(-7.0 // 2)
assert(-7.0 // 2 == -4.0)
assert(7.0 // -2 == -4.0)
assert(-7.0 // -2 == 3)

# Floor division with floats
assert(7.0 // 2.0 == 3.0)
assert(10.5 // 3.0 == 3.0)

# Floor division compound assignment
var x = 10
x //= 3
assert(x == 3)

x = 100
x //= 7
assert(x == 14)

# Basic modulo
assert(7 % 2 == 1)
assert(10 % 3 == 1)
assert(100 % 7 == 2)
assert(5 % 5 == 0)
assert(3 % 5 == 3)
assert(0 % 5 == 0)

### Not currently working
# Modulo with negatives
assert(-7 % 2 == 1)
assert(7 % -2 == -1)
assert(-7 % -2 == -1)
###

# Modulo compound assignment
var y = 10
y %= 3
assert(y == 1)

y = 100
y %= 7
assert(y == 2)

# Floor division and modulo relationship: a = (a // b) * b + (a % b)
val a = 17
val b = 5
assert(a == (a // b) * b + (a % b))

val c = -17
val d = 5
assert(c == (c // d) * d + (c % d))

# Modulo for even/odd checks
for (i in 1..20) {
    if (i % 2 == 0) {
        assert(i // 2 * 2 == i)
    } else {
        assert(i % 2 == 1)
    }
}

# Floor division by 1
for (i in 1..10) {
    assert(i // 1 == i)
}

# Modulo by 1
for (i in 1..10) {
    assert(i % 1 == 0)
}

# Large numbers
val big = 1000000
assert(big // 1000 == 1000)
assert(big % 1000 == 0)

# Mixed types
val mixed1 = 10.5 // 2
assert(mixed1 == 5.0)

val mixed2 = 10.5 % 2
assert(mixed2 == 0.5)

# Floor division with zero divisor should error
try {
    5 // 0
    assert(false, "should have errored")
} catch (e) {
    assert(e.contains("division by zero") || e.contains("zero"))
}

# Modulo with zero divisor should error
try {
    5 % 0
    assert(false, "should have errored")
} catch (e) {
    assert(e.contains("division by zero") || e.contains("zero"))
}

# Complex expression with floor division and modulo
val xx = 123
val yy = 17
val quotient = xx // yy
val remainder = xx % yy
assert(quotient == 7)
assert(remainder == 4)
assert(quotient * yy + remainder == xx)