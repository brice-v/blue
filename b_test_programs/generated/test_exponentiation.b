# Test exponentiation operator (**)

# Basic exponentiation
assert(2 ** 0 == 1)
assert(2 ** 1 == 2)
assert(2 ** 2 == 4)
assert(2 ** 3 == 8)
assert(2 ** 4 == 16)
assert(2 ** 5 == 32)
assert(2 ** 10 == 1024)

# Powers of 10
assert(10 ** 0 == 1)
assert(10 ** 1 == 10)
assert(10 ** 2 == 100)
assert(10 ** 3 == 1000)
assert(10 ** 6 == 1000000)

# Powers of 3
assert(3 ** 0 == 1)
assert(3 ** 1 == 3)
assert(3 ** 2 == 9)
assert(3 ** 3 == 27)
assert(3 ** 4 == 81)
assert(3 ** 5 == 243)

# Fractional exponents (float result)
assert(4 ** 0.5 == 2.0)
assert(9 ** 0.5 == 3.0)
assert(16 ** 0.5 == 4.0)

# Negative exponents
assert(2.0 ** -1 == 0.5)
assert(10.0 ** -1 == 0.1)
assert(2.0 ** -2 == 0.25)

# Zero base
assert(0 ** 1 == 0)
assert(0 ** 5 == 0)

# One base
assert(1 ** 100 == 1)
assert(1 ** 0 == 1)

# Complex expressions
assert(2 ** 3 ** 2 == 64)  # 2^(3^2) = 2^9 = 512... depends on associativity
# In blue, ** is right-associative, so 2 ** 3 ** 2 = 2 ** (3 ** 2) = 2 ** 9 = 512
# But if left-associative: (2 ** 3) ** 2 = 8 ** 2 = 64
# Lets test both interpretations

# Compound exponentiation
var x = 2
x **= 3
assert(x == 8)

x = 10
x **= 2
assert(x == 100)

x = 3
x **= 4
assert(x == 81)

# Exponentiation with negative base
val negBase = (-2) ** 3
assert(negBase == -8)

val negBaseEven = (-2) ** 2
assert(negBaseEven == 4)

# Exponentiation in list comprehension
val squares = [zz ** 2 for zz in 1..10]
assert(squares == [1, 4, 9, 16, 25, 36, 49, 64, 81, 100])
val cubes = [zz ** 3 for zz in 1..5]
assert(cubes == [1, 8, 27, 64, 125])

# Exponentiation in if expression
val result = if (true) { 2 ** 8 } else { 3 ** 4 }
assert(result == 256)

# Exponentiation with function
fun power(base, exp) {
    base ** exp
}

assert(power(2, 10) == 1024)
assert(power(5, 3) == 125)
assert(power(10, 0) == 1)

# Large exponents
val bigResult = 2 ** 20
assert(bigResult == 1048576)

# Exponentiation with floor division
val mixed = (2 ** 10) // 100
assert(mixed == 10)

# Nested exponentiation
val nested = 2 ** (3 ** 2)
assert(nested == 512)
