# Test while-style loops (for true with break/continue)

# Basic while loop
var counter = 0
for (true) {
    counter += 1
    if (counter >= 10) {
        break
    }
}
assert(counter == 10)

# While with continue
var evenSum = 0
var i = 0
for (true) {
    if (i >= 20) {
        break
    }
    i += 1
    if (i % 2 != 0) {
        continue
    }
    evenSum += i
}
assert(evenSum == 110)  # 2+4+6+8+10+12+14+16+18+20 = 110

# While loop with list processing
var items = [1, 2, 3, 4, 5]
var idx = 0
var sum = 0
for (true) {
    if (idx >= len(items)) {
        break
    }
    sum += items[idx]
    idx += 1
}
assert(sum == 15)

# While loop with string iteration
var text = "hello"
var charCount = 0
var ci = 0
for (true) {
    if (ci >= len(text)) {
        break
    }
    charCount += 1
    ci += 1
}
assert(charCount == 5)

# Nested while loops
var pairs = []
var ri = 0
for (true) {
    if (ri >= 3) {
        break
    }
    var cj = 0
    for (true) {
        if (cj >= 4) {
            break
        }
        pairs << [ri, cj]
        cj += 1
    }
    ri += 1
}
assert(len(pairs) == 12)

# While loop with early exit
var found = false
var search = [5, 3, 8, 1, 9, 4]
var si = 0
for (true) {
    if (si >= len(search)) {
        break
    }
    if (search[si] == 1) {
        found = true
        break
    }
    si += 1
}
assert(found == true)

# While loop with timeout pattern
var timeout = 0
var maxTimeout = 100
var done = false
for (true) {
    timeout += 1
    if (timeout >= maxTimeout) {
        done = true
        break
    }
    if (timeout == 50) {
        done = false
        break
    }
}
assert(done == false)
assert(timeout == 50)

# While loop accumulating results
var squares = []
var n = 1
for (true) {
    if (n > 10) {
        break
    }
    squares << n * n
    n += 1
}
assert(squares == [1, 4, 9, 16, 25, 36, 49, 64, 81, 100])

# While loop with map iteration
var myMap = {a: 1, b: 2, c: 3}
var keys = myMap.keys()
var mi = 0
var mapSum = 0
for (true) {
    if (mi >= len(keys)) {
        break
    }
    mapSum += myMap[keys[mi]]
    mi += 1
}
assert(mapSum == 6)

# While loop with break in nested block
var result = 0
for (true) {
    var x = 5
    if (x > 3) {
        result = x * 2
        break
    }
}
assert(result == 10)

# While loop with continue in nested block
var count = 0
for (true) {
    count += 1
    if (count >= 5) {
        break
    }
    {
        if (count == 3) {
            continue
        }
    }
}
assert(count == 5)

# While loop that doesnt execute
var neverExecuted = 0
for (false) {
    neverExecuted += 1
}
assert(neverExecuted == 0)

# While loop with complex condition
var accumulator = 0
var step = 1
for (true) {
    if (accumulator >= 100) {
        break
    }
    accumulator += step
    step += 1
}
# 1+2+3+...+14 = 105 >= 100, so 14 steps
assert(accumulator == 105)
assert(step == 15)  # step was incremented one more time