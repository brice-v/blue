# Test list comprehensions with multiple generators, filters, and edge cases

# Basic list comprehension
val squares = [x * x for x in 1..5]
assert(squares == [1, 4, 9, 16, 25])

# List comprehension with filter
val evens = [x for x in 1..10 if x % 2 == 0]
assert(evens == [2, 4, 6, 8, 10])

### Not yet supported
# Multiple generators
val pairs = [x + y for x in 1..3 for y in 1..3]
assert(len(pairs) == 9)
assert(pairs == [2, 3, 4, 3, 4, 5, 4, 5, 6])

# Multiple generators with filter
val oddPairs = [x + y for x in 1..5 for y in 1..5 if (x + y) % 2 == 1]
# Sum is odd when one is even and one is odd
# (1,2),(1,4),(3,2),(3,4),(5,2),(5,4),(2,1),(2,3),(2,5),(4,1),(4,3),(4,5) = 12 pairs
assert(len(oddPairs) == 12)

# Nested comprehension with variables
val grid = [[x * 10 + y for y in 1..3] for x in 1..3]
assert(len(grid) == 3)
assert(grid[0] == [11, 12, 13])
assert(grid[1] == [21, 22, 23])
assert(grid[2] == [31, 32, 33])
###

# List comprehension with list concatenation
val doubled = [x * 2 for x in [1, 2, 3, 4, 5]]
assert(doubled == [2, 4, 6, 8, 10])

# List comprehension with string
val upperChars = [c.to_upper() for c in "hello"]
assert(upperChars == ["H", "E", "L", "L", "O"])

# List comprehension with condition
val filtered = [x for x in 1..20 if x % 3 == 0 or x % 5 == 0]
assert(filtered == [3, 5, 6, 9, 10, 12, 15, 18, 20])

# List comprehension with function call
fun double(x) { x * 2 }
val result = [double(x) for x in 1..5]
assert(result == [2, 4, 6, 8, 10])

# Empty list comprehension
val empty = [x for x in []]
assert(empty == [])

# List comprehension with range
val rangeSq = [x ** 2 for x in 0..<10]
assert(len(rangeSq) == 10)
assert(rangeSq[0] == 0)
assert(rangeSq[9] == 81)

# List comprehension with nested function calls
fun add(a, b) { a + b }
val mixed = [add(x, y) for x in 1..3 for y in 4..6]
assert(len(mixed) == 9)
assert(mixed[0] == 5)  # 1+4
assert(mixed[8] == 9)  # 3+6

# List comprehension with boolean expression
val bools = [x > 5 for x in 1..10]
assert(bools == [false, false, false, false, false, true, true, true, true, true])

# List comprehension with ternary-like if
val categorized = [if (x % 2 == 0) { "even" } else { "odd" } for x in 1..6]
assert(categorized == ["odd", "even", "odd", "even", "odd", "even"])

# List comprehension with map access
val data = {a: 1, b: 2, c: 3}
val vals = [data[k] for k in ["a", "c"]]
assert(vals == [1, 3])

# List comprehension with set membership
val nums = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
val primes = [x for x in nums if x in {2, 3, 5, 7}]
assert(primes == [2, 3, 5, 7])

# List comprehension with string operations
val words = ["hello", "world", "blue"]
val lengths = [w.len() for w in words]
assert(lengths == [5, 5, 4])

# List comprehension with type conversion
val strNums = [str(x) for x in 1..5]
assert(strNums == ["1", "2", "3", "4", "5"])

# C-style for loop in comprehension
val cStyle = [i for (var i = 0; i < 5; i += 1) if i > 0]
assert(cStyle == [1, 2, 3, 4])

# List comprehension with closure
val maker = fun(multiplier) {
    [x * multiplier for x in 1..5]
}
val times3 = maker(3)
assert(times3 == [3, 6, 9, 12, 15])

# List comprehension with complex expression
val complex = [x ** 2 + y for x in 1..3 for y in 1..3 if x + y <= 4]
# x=1: y=1,2,3 (all <= 4): 1+1=2, 1+2=3, 1+3=4
# x=2: y=1,2 (2+1=3, 2+2=4): 4+1=5, 4+2=6
# x=3: y=1 (3+1=4): 9+1=10
assert(complex == [2, 3, 4, 5, 6, 10])

# Multiple clauses with if on first clause
val ifThenFor = [x for x in 1..5 if x > 2 for y in 1..3]
assert(len(ifThenFor) == 9)
assert(ifThenFor == [3, 3, 3, 4, 4, 4, 5, 5, 5])

# Multiple clauses with filter on inner value
val innerFilter = [y for x in 1..3 if x > 2 for y in 1..3]
assert(innerFilter == [1, 2, 3])

# Flatten list of lists
val nested = [[1,2],[3,4],[5,6]]
val flat = [item for sublist in nested for item in sublist]
assert(flat == [1, 2, 3, 4, 5, 6])