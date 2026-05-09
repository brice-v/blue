# Test set operations: union, intersection, difference, symmetric difference

# Basic set creation
val s1 = {1, 2, 3}
val s2 = {3, 4, 5}

# Union (|)
val union = s1 | s2
assert(union == {1, 2, 3, 4, 5})

# Intersection (&)
val intersection = s1 & s2
assert(intersection == {3})

# Difference (-)
val diff = s1 - s2
assert(diff == {1, 2})

val diff2 = s2 - s1
assert(diff2 == {4, 5})

# Symmetric difference (^)
val symDiff = s1 ^ s2
assert(symDiff == {1, 2, 4, 5})

# Empty set operations
val empty = set([])
val s3 = {1, 2, 3}

val unionEmpty = empty | s3
assert(unionEmpty == {1, 2, 3})

val intersectEmpty = empty & s3
assert(intersectEmpty == set([]))

val diffEmpty = s3 - empty
assert(diffEmpty == {1, 2, 3})

# Subset and superset
val subset = {1, 2}
val superset = {1, 2, 3, 4}

assert(subset <= superset)   # subset is subset of superset
assert(superset >= subset)   # superset is superset of subset
assert(!(subset >= superset))  # subset is not superset of superset
assert(!(superset <= subset))  # superset is not subset of subset

# Equal sets
assert({1, 2, 3} == {3, 2, 1})  # order doesnt matter
assert({1, 2, 3} != {1, 2, 3, 4})

# Set with strings
val strSet1 = {"a", "b", "c"}
val strSet2 = {"c", "d", "e"}

assert(strSet1 | strSet2 == {"a", "b", "c", "d", "e"})
assert(strSet1 & strSet2 == {"c"})
assert(strSet1 - strSet2 == {"a", "b"})
assert(strSet1 ^ strSet2 == {"a", "b", "d", "e"})

# Set with mixed types
val mixedSet = {1, "hello", true}
assert(len(mixedSet) == 3)

# Set membership
val numSet = {1, 2, 3, 4, 5}
assert(3 in numSet)
assert(6 notin numSet)
assert(1 in numSet)
assert(0 notin numSet)

# Set operations with single element
val single = {42}
val other = {1, 42, 99}

assert(single | other == {1, 42, 99})
assert(single & other == {42})
assert(other - single == {1, 99})

# Set from list
val fromList = set([1, 2, 2, 3, 3, 3])
assert(fromList == {1, 2, 3})  # duplicates removed

# Set to list
val asList = fromList.to_list()
assert(len(asList) == 3)
assert(1 in set(asList))
assert(2 in set(asList))
assert(3 in set(asList))

# Set with boolean values
val boolSet = {true, false}
assert(len(boolSet) == 2)
assert(true in boolSet)
assert(false in boolSet)

# Set with null
val nullSet = {null, 1, 2}
assert(len(nullSet) == 3)
assert(null in nullSet)

# Chained set operations
val a = {1, 2, 3, 4, 5}
val b = {3, 4, 5, 6, 7}
val c = {5, 6, 7, 8, 9}

# (a | b) - c = {1, 2, 3, 4, 5, 6, 7} - {5, 6, 7, 8, 9} = {1, 2, 3, 4}
val chained = (a | b) - c
assert(chained == {1, 2, 3, 4})

# a - (b & c) = {1, 2, 3, 4, 5} - {5, 6, 7} = {1, 2, 3, 4}
val chained2 = a - (b & c)
assert(chained2 == {1, 2, 3, 4})

# Set comprehension
val evenSet = {x for x in 1..10 if x % 2 == 0}
assert(evenSet == {2, 4, 6, 8, 10})

val squareSet = {x * x for x in 1..5}
assert(squareSet == {1, 4, 9, 16, 25})

# Set with complex expressions
val bigSet = {x + y for x in {1, 2} for y in {3, 4}}
assert(bigSet == {4, 5, 6})

# Empty set operations
val emptySet = set([])
assert(len(emptySet) == 0)
assert(emptySet <= {1, 2, 3})  # empty set is subset of everything
assert(!({1, 2, 3} <= emptySet))  # non-empty is not subset of empty

# Set operations with ranges
val rangeSet = {x for x in 1..5}
assert(rangeSet == {1, 2, 3, 4, 5})
