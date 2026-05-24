# Test C-style for loops: for (var i = 0; condition; increment)

# Basic C-style for loop
var sum = 0
for (var i = 0; i < 10; i += 1) {
    sum += i
}
assert(sum == 45)  # 0+1+2+3+4+5+6+7+8+9 = 45

# C-style for with decrement
var count = 0
for (var i = 10; i > 0; i -= 1) {
    count += 1
}
assert(count == 10)

# C-style for with step of 2
var total = 0
for (var i = 0; i < 20; i += 2) {
    total += i
}
assert(total == 90)  # 0+2+4+6+8+10+12+14+16+18 = 90

# C-style for with multiplication step
var powers = []
for (var i = 1; i <= 1024; i *= 2) {
    powers << i
}
assert(len(powers) == 11)  # 1,2,4,8,16,32,64,128,256,512,1024
assert(powers[0] == 1)
assert(powers[10] == 1024)

# C-style for with break
var found = null
for (var i = 0; i < 100; i += 1) {
    if (i == 42) {
        found = i
        break
    }
}
assert(found == 42)

# C-style for with continue
var oddCount = 0
for (var i = 0; i < 20; i += 1) {
    if (i % 2 == 0) {
        continue
    }
    oddCount += 1
}
assert(oddCount == 10)

# C-style for with nested loop
var pairs = []
for (var i = 0; i < 3; i += 1) {
    for (var j = 0; j < 4; j += 1) {
        pairs << [i, j]
    }
}
assert(len(pairs) == 12)

# C-style for with variable in condition
var iterations = 0
for (var i = 0; i < 5; i += 1) {
    iterations += 1
}
assert(iterations == 5)

# C-style for with empty body
var result = 0
for (var i = 0; i < 5; i += 1) {
    result += i
}
assert(result == 10)  # 0+1+2+3+4 = 10

# C-style for with complex condition
var matches = 0
for (var i = 1; i <= 50; i += 1) {
    if (i > 10 && i <= 30 && i % 3 == 0) {
        matches += 1
    }
}
# Numbers in range (11..30] divisible by 3: 12,15,18,21,24,27,30 = 7
assert(matches == 7)

# C-style for with float comparison
var floatSum = 0.0
for (var i = 0.0; i < 5.0; i += 1.0) {
    floatSum += i
}
assert(floatSum == 10.0)  # 0+1+2+3+4 = 10

# Infinite loop with break
var counter = 0
for (true) {
    counter += 1
    if (counter >= 100) {
        break
    }
}
assert(counter == 100)

# C-style for with function call in condition
fun isLessThan(v, max) {
    v < max
}

var count2 = 0
for (var i = 0; isLessThan(i, 5); i += 1) {
    count2 += 1
}
assert(count2 == 5)

# C-style for with list index
var list = [10, 20, 30, 40, 50]
var sum4 = 0
for (var i = 0; i < len(list); i += 1) {
    sum4 += list[i]
}
assert(sum4 == 150)
