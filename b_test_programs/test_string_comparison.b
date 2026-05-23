# Test string comparison operators: > >= < <=

# Less than
assert("a" < "b")
assert("a" < "a" == false)
assert("b" < "a" == false)
assert("a" < "ab")
assert("ab" < "a" == false)
assert("abc" < "abd")
assert("abc" < "abc" == false)
println("Less than: PASS")

# Less than or equal
assert("a" <= "b")
assert("a" <= "a")
assert("b" <= "a" == false)
assert("abc" <= "abd")
assert("abc" <= "abc")
println("Less than or equal: PASS")

# Greater than
assert("b" > "a")
assert("a" > "a" == false)
assert("a" > "b" == false)
assert("ab" > "a")
assert("a" > "ab" == false)
println("Greater than: PASS")

# Greater than or equal
assert("b" >= "a")
assert("a" >= "a")
assert("a" >= "b" == false)
assert("abc" >= "abc")
assert("abd" >= "abc")
assert("abc" >= "abd" == false)
println("Greater than or equal: PASS")

# Empty string comparisons
assert("" < "a")
assert("a" > "")
assert("" <= "")
assert("" >= "")
println("Empty string: PASS")

# Case sensitivity
assert("A" < "a")  # uppercase before lowercase in ASCII
assert("Z" < "a")
assert("a" > "Z")
println("Case sensitivity: PASS")

# Single character
assert("z" > "a")
assert("0" < "9")
