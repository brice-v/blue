# Comprehensive test for list/set/map comprehensions with multiple for/if clauses
# Designed to mirror Pythons comprehension semantics

println("=== List Comprehensions ===")

# Basic list comprehension
val squares = [x * x for x in 1..5]
assert(squares == [1, 4, 9, 16, 25])

# List comprehension with filter
val evens = [x for x in 1..10 if x % 2 == 0]
assert(evens == [2, 4, 6, 8, 10])

# Nested for clauses (cartesian product)
val pairs = [x + y for x in 1..3 for y in 1..3]
assert(pairs == [2, 3, 4, 3, 4, 5, 4, 5, 6])

# Nested for clauses with filter on last for
val oddPairs = [x + y for x in 1..5 for y in 1..5 if (x + y) % 2 == 1]
assert(len(oddPairs) == 12)

# Filter on first for clause (if before a for)
val ifThenFor = [x for x in 1..5 if x > 2 for y in 1..3]
assert(len(ifThenFor) == 9)
assert(ifThenFor == [3, 3, 3, 4, 4, 4, 5, 5, 5])

# Filter on first for, binding inner value
val innerFilter = [y for x in 1..3 if x > 2 for y in 1..3]
assert(innerFilter == [1, 2, 3])

# Flatten list of lists (Python: [item for sublist in nested for item in sublist])
val nested = [[1,2],[3,4],[5,6]]
val flat = [item for sublist in nested for item in sublist]
assert(flat == [1, 2, 3, 4, 5, 6])

# Use x in set for filter condition
val nums = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
val primes = [x for x in nums if x in {2, 3, 5, 7}]
assert(primes == [2, 3, 5, 7])

# Use x notin set for filter condition  
val notPrimes = [x for x in nums if x notin {2, 3, 5, 7}]
assert(notPrimes == [1, 4, 6, 8, 9, 10])

# Filter on both for clauses
val filtered = [x + y for x in 1..5 if x % 2 == 1 for y in 1..5 if y % 2 == 1]
# x odd: 1,3,5; y odd: 1,3,5 => 9 pairs
assert(len(filtered) == 9)
assert(filtered == [2, 4, 6, 4, 6, 8, 6, 8, 10])

# Empty input
val empty = [x for x in []]
assert(empty == [])

# Empty input with filter
val empty2 = [x for x in 1..5 if x > 10]
assert(empty2 == [])

# Nested for with empty inner
val empty3 = [x + y for x in 1..3 for y in []]
assert(empty3 == [])

# Complex expression
val complex = [x ** 2 + y for x in 1..3 for y in 1..3 if x + y <= 4]
assert(complex == [2, 3, 4, 5, 6, 10])

# Function call in filtering
fun isOdd(n) { n % 2 == 1 }
val odds = [x for x in 1..10 if isOdd(x)]
assert(odds == [1, 3, 5, 7, 9])

# Four for clauses
val fourWay = [x + y + z + w for x in 1..2 for y in 1..2 for z in 1..2 for w in 1..2]
assert(len(fourWay) == 16)
# First element: 1+1+1+1 = 4, Last: 2+2+2+2 = 8
assert(fourWay[0] == 4)
assert(fourWay[15] == 8)

println("  list: OK")

println("=== Set Comprehensions ===")

# Basic set comprehension
val setSquares = {x * x for x in 1..5}
assert(setSquares == {1, 4, 9, 16, 25})

# Set comprehension with filter
val setEvens = {x for x in 1..10 if x % 2 == 0}
assert(setEvens == {2, 4, 6, 8, 10})

# Nested for clauses with set (deduplicates)
val setPairs = {x + y for x in 1..3 for y in 1..3}
assert(setPairs == {2, 3, 4, 5, 6})
# Python: set([1+1,1+2,1+3,2+1,2+2,2+3,3+1,3+2,3+3]) = {2,3,4,5,6}

# Nested for with filter on first for
val setIfThen = {x for x in 1..5 if x > 2 for y in 1..3}
assert(setIfThen == {3, 4, 5})

# Set comprehension with in filter
val setPrimes = {x for x in nums if x in {2, 3, 5, 7}}
assert(setPrimes == {2, 3, 5, 7})

println("  set: OK")

println("=== Map (Dict) Comprehensions ===")

# Basic map comprehension
val mapSquares = {x: x * x for x in 1..5}
assert(mapSquares == {1: 1, 2: 4, 3: 9, 4: 16, 5: 25})

# Map comprehension with filter
val mapEvens = {x: x for x in 1..10 if x % 2 == 0}
assert(mapEvens == {2: 2, 4: 4, 6: 6, 8: 8, 10: 10})

# Nested for clauses with map (unique key via x*10+y)
val mapPairs = {x * 10 + y: x * y for x in 1..3 for y in 1..3}
assert(len(mapPairs) == 9)
assert(mapPairs[11] == 1)  # x=1,y=1 => key=11, val=1
assert(mapPairs[33] == 9)  # x=3,y=3 => key=33, val=9
assert(mapPairs[12] == 2)  # x=1,y=2 => key=12, val=2

# Nested for with filter on first for
val mapIfThen = {x: x + y for x in 1..5 if x > 2 for y in 1..3}
assert(len(mapIfThen) == 3)
assert(mapIfThen[3] == 6)  # x=3, y=3: 3+3=6 (last y for x=3)
assert(mapIfThen[4] == 7)  # x=4, y=3: 4+3=7
assert(mapIfThen[5] == 8)  # x=5, y=3: 5+3=8

# Map comprehension with in filter
val m = {x: x * 10 for x in 1..10 if x notin {1, 3, 5, 7, 9}}
assert(m == {2: 20, 4: 40, 6: 60, 8: 80, 10: 100})

println("  map: OK")

println("=== Edge Cases ===")

# Single element
val single = [x for x in 1..1]
assert(single == [1])

# Range to end
val endRange = [x for x in 0..<5]
assert(endRange == [0, 1, 2, 3, 4])

# Non-integer types (strings)
val chars = [c for c in "abc"]
assert(chars == ["a", "b", "c"])

# Filter with complex boolean expression
val filtered2 = [x for x in 1..20 if x % 3 == 0 or x % 5 == 0]
assert(filtered2 == [3, 5, 6, 9, 10, 12, 15, 18, 20])

# Nested for with string
val charPairs = [a + b for a in "ab" for b in "xy"]
assert(charPairs == ["ax", "ay", "bx", "by"])

# Comprehension inside function call
fun getEvens(limit) {
    [x for x in 1..limit if x % 2 == 0]
}
val fnResult = getEvens(10)
assert(fnResult == [2, 4, 6, 8, 10])

# Comprehension with closure capturing outer variable
fun multiplier(factor) {
    [x * factor for x in 1..5]
}
val times3 = multiplier(3)
assert(times3 == [3, 6, 9, 12, 15])
val times5 = multiplier(5)
assert(times5 == [5, 10, 15, 20, 25])

println("  edge cases: OK")

println("ALL COMPREHENSION TESTS PASSED")
