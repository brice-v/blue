# Test edge cases and VM-specific features

# Empty list operations
val emptyList = []
assert(len(emptyList) == 0)
assert(type(emptyList) == "LIST")

# Empty map operations
val emptyMap = {}
assert(len(emptyMap) == 0)
assert(type(emptyMap) == "MAP")

# Empty set operations
val emptySet = set([])
assert(len(emptySet) == 0)
assert(type(emptySet) == "SET")

# Nil operations
assert(null == null)
assert(type(null) == "NULL")

# Boolean operations
assert(true == true)
assert(false == false)
assert(true != false)
assert(not true == false)
assert(not false == true)

# Truthy/falsy
assert(!([]))  # empty list is falsy
assert(!({}))  # empty map is falsy
assert(!set([]))  # empty set is falsy
assert(!!(""))  # empty string is truthy
assert(!(null))  # null is falsy
assert(!(false))  # false is falsy
assert(!!("hello"))  # non-empty string is truthy
assert(!!([1]))  # non-empty list is truthy
assert(!!({a: 1}))  # non-empty map is truthy
assert(!!(1))  # non-zero number is truthy
assert(!!(0))  # zero is truthy

# Range operations
val range1 = 1..5
val range2 = 1..<5

# Range in for loop
var count = 0
for (i in 1..5) {
    count += 1
}
assert(count == 5)

count = 0
for (i in 1..<5) {
    count += 1
}
assert(count == 4)

# Range with step (via C-style loop)
var stepCount = 0
for (var i = 0; i < 10; i += 2) {
    stepCount += 1
}
assert(stepCount == 5)  # 0,2,4,6,8 = 5 iterations

# Large number operations
val bigNum = 2 ** 64
assert(bigNum > 0)

# Float precision
val piApprox = 3.14159
assert(piApprox > 3.14)
assert(piApprox < 3.15)

# Division edge cases
val half = 1.0 / 2
assert(half == 0.5)

val zeroDiv = 0 / 5
assert(zeroDiv == 0)

# Modulo edge cases
val modZero = 5 % 1
assert(modZero == 0)

val modSelf = 5 % 5
assert(modSelf == 0)

val modSmaller = 3 % 5
assert(modSmaller == 3)

# Negation edge cases
val negZero = -0
assert(negZero == 0)

val negNeg = --5
assert(negNeg == 5)

# String edge cases
assert("".len() == 0)
assert("a".len() == 1)
assert("hello" + "" == "hello")
assert("" + "world" == "world")
assert("" * 5 == "")
assert(5 * "" == "")

# List edge cases
val single = [1]
assert(single[0] == 1)
assert(single.len() == 1)

val nested = [[1, 2], [3, 4]]
assert(nested[0][1] == 2)
assert(nested[1][0] == 3)

# Map edge cases
val deepMap = {a: {b: {c: 3}}}
assert(deepMap["a"]["b"]["c"] == 3)

# Set edge cases
val singleSet = {1}
assert(len(singleSet) == 1)
assert(1 in singleSet)

# Comparison edge cases
assert(0 == 0.0)
assert(1 == 1.0)
assert("a" == "a")
assert(true == true)
assert(false == false)

# Type conversion edge cases
assert(int("123") == 123)
assert(int(456.7) == 456)
assert(float(123) == 123.0)
assert(str(42) == "42")
assert(str(3.14) == "3.14")

# List concatenation edge cases
assert([] + [] == [])
assert([1] + [] == [1])
assert([] + [1] == [1])
assert([1, 2] + [3, 4] == [1, 2, 3, 4])

# Map merging via comprehension
val m1 = {a: 1, b: 2}
val m2 = {c: 3, d: 4}
#val merged = {k: v for [k, v] in m1} + {k: v for [k, v] in m2}
# Note: this depends on how + works for maps

# List repetition
assert("ab" * 0 == "")
assert("ab" * 1 == "ab")
assert("ab" * 3 == "ababab")

# Floor division edge cases
assert(0 // 5 == 0)
assert(5 // 1 == 5)
assert(1 // 5 == 0)

# Exponentiation edge cases
assert(1 ** 100 == 1)
assert(0 ** 0 == 1)  # 0^0 = 1 in most languages
assert(2 ** 0 == 1)
assert(0 ** 1 == 0)

### Not yet supported
# Bitwise edge cases
assert((0 & 0) == 0)
assert((0 | 0) == 0)
assert((0 ^ 0) == 0)
assert((~0) == -1)
assert((0 << 5) == 0)
assert((0 >> 5) == 0)
###

# String interpolation edge cases
val emptyInterp = "value: #{0}"
assert(emptyInterp == "value: 0")

val nestedInterp = "a: #{1 + 2}"
assert(nestedInterp == "a: 3")

# Lambda edge cases
val emptyLambda = |x| => x
assert(emptyLambda(5) == 5)

val multiParamLambda = |a, b, c| => a + b + c
assert(multiParamLambda(1, 2, 3) == 6)

### Not yet supported
# Match edge cases
val noMatch = match (999) {
    1 => { "one" },
}
assert(noMatch == null)
###

val matchAll = match (5) {
    _ => { "anything" },
}
assert(matchAll == "anything")

# Assert edge cases
assert(1 == 1)
assert(true)
assert(false == false)
assert(null == null)
assert([] == [])
assert({} == {})

# Self edge cases
val myPid = self()
assert(type(myPid) == "PROCESS")
assert(myPid.id >= 0)

# Eval edge cases
val evalResult = eval("1 + 1")
assert(evalResult == 2)

# Type edge cases
assert(type(1) == "INTEGER")
assert(type(1.0) == "FLOAT")
assert(type("hello") == "STRING")
assert(type(true) == "BOOLEAN")
assert(type(null) == "NULL")
assert(type([]) == "LIST")
assert(type({}) == "MAP")
assert(type(set([])) == "SET")
