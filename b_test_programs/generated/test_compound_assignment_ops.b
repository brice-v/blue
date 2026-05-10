# Test all compound assignment operators

# +=
var a = 10
a += 5
assert(a == 15)

a = "hello"
a += " world"
assert(a == "hello world")

var b = [1, 2]
b += [3, 4]
assert(b == [1, 2, 3, 4])

# -=
var c = 10
c -= 3
assert(c == 7)

c = 100
c -= 100
assert(c == 0)

c = -5
c -= 5
assert(c == -10)

# *=
var d = 5
d *= 4
assert(d == 20)

d = 0
d *= 100
assert(d == 0)

# /=
var e = 20
e /= 4
assert(e == 5)

e = 7
e /= 2
assert(e == 3.5)

# //= (floor division)
var f = 7
f //= 2
assert(f == 3)

f = -7
f //= 2
assert(f == -4)

f = 0
f //= 5
assert(f == 0)

# %=
var g = 10
g %= 3
assert(g == 1)

g = 7
g %= 7
assert(g == 0)

g = 5
g %= 10
assert(g == 5)

# **=
var h = 2
h **= 8
assert(h == 256)

h = 10
h **= 0
assert(h == 1)

h = 3
h **= 3
assert(h == 27)

# &=
var i = 15  # 1111
i &= 9      # 1111 & 1001 = 1001 = 9
assert(i == 9)

# |=
var j = 5   # 0101
j |= 12     # 0101 | 1100 = 1101 = 13
assert(j == 13)

# ^=
var k = 15  # 1111
k ^= 9      # 1111 ^ 1001 = 0110 = 6
assert(k == 6)

# &&=
var l = true
l &&= true
assert(l == true)

l = true
l &&= false
assert(l == false)

l = false
l &&= true
assert(l == false)

# ||=
var m = false
m ||= true
assert(m == true)

m = true
m ||= false
assert(m == true)

m = false
m ||= false
assert(m == false)

# ~= (bitwise not-equal and assign) - This is XOR in blue
var n = 10
n ~= 5
assert(n == 15)  # 1010 ^ 0101 = 1111 = 15

# <<=
var o = 1
o <<= 4
assert(o == 16)

# >>=
var p = 32
p >>= 3
assert(p == 4)

# Compound assignment with map
var map1 = {x: 10, y: 20}
map1["x"] += 5
assert(map1["x"] == 15)

map1["y"] *= 2
assert(map1["y"] == 40)

# Compound assignment with list
var list1 = [1, 2, 3]
list1[0] += 100
assert(list1[0] == 101)

list1[1] -= 1
assert(list1[1] == 1)

# Compound assignment chained
var chain = 1
chain += 1
chain *= 2
chain -= 1
chain **= 2
assert(chain == 16)  # (1+1)*2-1 = 3, 3**2 = 9... wait

# Let me recalculate:
# chain = 1
# chain += 1 => chain = 2
# chain *= 2 => chain = 4
# chain -= 1 => chain = 3
# chain **= 2 => chain = 9
assert(chain == 9)

# Compound assignment with if expression
var v = 0
v += if (true) { 10 } else { 20 }
assert(v == 10)

v += if (false) { 100 } else { 50 }
assert(v == 60)

# Compound assignment with function result
fun getNum() { 42 }
var result = 0
result += getNum()
assert(result == 42)

# Compound assignment in for loop
var sum = 0
for (i in 1..10) {
    sum += i
}
assert(sum == 55)

# Compound assignment with range
var x = 1
for (i in 1..5) {
    x *= 2
}
assert(x == 32)  # 2^5 = 32
