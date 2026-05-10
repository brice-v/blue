# Test operator precedence

# Multiplication before addition
assert(2 + 3 * 4 == 14)  # 2 + (3*4) = 14
assert(10 - 2 * 3 == 4)  # 10 - (2*3) = 4

# Exponentiation before multiplication
assert(2 + 3 ** 2 == 11)  # 2 + (3**2) = 11
assert(2 * 3 ** 2 == 18)  # 2 * (3**2) = 18

# Negation before multiplication
assert(-2 * 3 == -6)  # (-2) * 3 = -6

# Comparison after arithmetic
assert(1 + 2 < 5)
assert(10 - 3 > 5)
assert(2 + 2 == 4)
assert(2 + 3 != 6)

# AND before OR
assert(true or false and false == true)  # true or (false and false) = true
assert((false or false and true == false) == false)  # false or (false and true) = false

# NOT before AND
assert(not false and true == true)  # (not false) and true = true
assert((not true and false == false) == false)  # (not true) and false = false

# Parenthesized expressions override precedence
assert((2 + 3) * 4 == 20)  # (2+3) * 4 = 20
assert(2 + (3 * 4) == 14)  # 2 + (3*4) = 14

# Complex expression
val complex = 1 + 2 * 3 ** 2 - 4 / 2
# = 1 + (2 * 9) - (4 / 2)
# = 1 + 18 - 2
# = 17
assert(complex == 17)

# Bitwise AND before OR
val bitwise = 0u1 | 0u2 & 0u4
# 1 | (2 & 4) = 1 | 0 = 1
assert(bitwise == 0u1)

# Range before in
val inResult = 5 in 1..10
assert(inResult == true)

# Not before in
val notInResult = not (5 in 1..3)
# not (5 in 1..3) = not false = true
assert(notInResult == true)

# Comparison with equality
val cmp = 1 + 1 == 2
assert(cmp == true)

# Chained comparisons (not supported in most languages, but lets test)
val chained = 1 < 2 and 2 < 3
assert(chained == true)

# Function call after index
val list = [1, 2, 3]
val indexResult = list[0] + 1
assert(indexResult == 2)

# Member access after call
val strResult = "hello".to_upper()
assert(strResult == "HELLO")

# Index with expression
val list2 = [0, 10, 20, 30]
val exprIndex = list2[1 + 1]
assert(exprIndex == 20)

# Multiple unary operators
val doubleNeg = --5  # -(-5) = 5
assert(doubleNeg == 5)

val notNot = not not true
assert(notNot == true)

# Mix of arithmetic and comparison
val mix = 10 > 5 + 3
# 10 > (5 + 3) = 10 > 8 = true
assert(mix == true)

# Mix of logical and comparison
val logicMix = 5 > 3 and 2 < 4
assert(logicMix == true)

# Exponentiation precedence with negative base
val negPow = (-2) ** 3
assert(negPow == -8)

# Floor division vs regular division
val floorVsDiv = 7 // 2
assert(floorVsDiv == 3)
val regularDiv = 7.0 / 2
assert(regularDiv == 3.5)

# Modulo with floor division
val modResult = 7 % 2
assert(modResult == 1)

# Compound assignment precedence
var x = 1
x += 2 * 3
assert(x == 7)  # x = 1 + (2*3) = 7

var y = 10
y -= 3 ** 2
assert(y == 1)  # y = 10 - 9 = 1

### Not yet supported
# Shift vs addition
val shiftAdd = 0u1 << 0u2 + 0u1
# 1 << (2 + 1) = 1 << 3 = 8
println(shiftAdd)
assert(shiftAdd == 0u8)

# Shift vs comparison
val shiftCmp = 0u1 << 0u2 > 0u3
# (1 << 2) > 3 = 4 > 3 = true
assert(shiftCmp)
###

# String concat vs arithmetic
val strMix = "a" + "b" * 2
# "a" + ("b" * 2) = "a" + "bb" = "abb"
assert(strMix == "abb")