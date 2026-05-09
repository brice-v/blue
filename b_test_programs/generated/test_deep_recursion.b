# Test deep recursion

# Factorial
fun factorial(n) {
    if (n <= 1) {
        return 1
    }
    return n * factorial(n - 1)
}

assert(factorial(0) == 1)
assert(factorial(1) == 1)
assert(factorial(5) == 120)
assert(factorial(10) == 3628800)

# Fibonacci
fun fib(n) {
    if (n <= 0) {
        return 0
    }
    if (n == 1) {
        return 1
    }
    return fib(n - 1) + fib(n - 2)
}

assert(fib(0) == 0)
assert(fib(1) == 1)
assert(fib(10) == 55)
assert(fib(20) == 6765)

# Ackermann-like function (simplified)
fun ackermann(m, n) {
    if (m == 0) {
        return n + 1
    }
    if (n == 0) {
        return ackermann(m - 1, 1)
    }
    return ackermann(m - 1, ackermann(m, n - 1))
}

assert(ackermann(0, 0) == 1)
assert(ackermann(1, 0) == 2)
assert(ackermann(1, 1) == 3)
assert(ackermann(2, 2) == 7)

# Deep recursion with single call
fun deep(n) {
    if (n <= 0) {
        return 0
    }
    return 1 + deep(n - 1)
}

assert(deep(100) == 100)
assert(deep(200) == 200)

# Mutual recursion
fun isEven(n) {
    if (n == 0) {
        return true
    }
    return isOdd(n - 1)
}

fun isOdd(n) {
    if (n == 0) {
        return false
    }
    return isEven(n - 1)
}

assert(isEven(0) == true)
assert(isEven(1) == false)
assert(isEven(2) == true)
assert(isEven(10) == true)
assert(isOdd(0) == false)
assert(isOdd(1) == true)
assert(isOdd(2) == false)
assert(isOdd(10) == false)

# Tree traversal via recursion
fun buildTree(depth) {
    if (depth <= 0) {
        return []
    }
    return [depth] + buildTree(depth - 1)
}

val tree = buildTree(5)
assert(tree == [5, 4, 3, 2, 1])

# Tree sum
fun treeSum(depth) {
    if (depth <= 0) {
        return 0
    }
    return depth + treeSum(depth - 1)
}

assert(treeSum(10) == 55)  # 1+2+3+...+10 = 55

# Recursive list processing
fun listSum(lst) {
    if (len(lst) == 0) {
        return 0
    }
    return lst[0] + listSum(lst[1..])
}

assert(listSum([1, 2, 3, 4, 5]) == 15)
assert(listSum([]) == 0)
assert(listSum([10]) == 10)

# Recursive list length
fun listLen(lst) {
    if (len(lst) == 0) {
        return 0
    }
    return 1 + listLen(lst[1..])
}

assert(listLen([1, 2, 3]) == 3)
assert(listLen([]) == 0)

# Recursive map processing
fun mapSum(m) {
    var sum = 0
    for ([k, v] in m) {
        sum += v
    }
    return sum
}

val testMap = {a: 1, b: 2, c: 3}
assert(mapSum(testMap) == 6)

# Recursive string processing
fun reverseStr(s) {
    if (len(s) <= 1) {
        return s
    }
    return reverseStr(s[1..]) + s[0]
}

assert(reverseStr("hello") == "olleh")
assert(reverseStr("abc") == "cba")
assert(reverseStr("") == "")
assert(reverseStr("a") == "a")

# Recursive power
fun power(base, exp) {
    if (exp == 0) {
        return 1
    }
    if (exp == 1) {
        return base
    }
    if (exp % 2 == 0) {
        val half = power(base, exp / 2)
        return half * half
    }
    return base * power(base, exp - 1)
}

assert(power(2, 0) == 1)
assert(power(2, 1) == 2)
assert(power(2, 10) == 1024)
assert(power(3, 4) == 81)
assert(power(10, 3) == 1000)

# Recursive GCD
fun gcd(a, b) {
    if (b == 0) {
        return a
    }
    return gcd(b, a % b)
}

assert(gcd(12, 8) == 4)
assert(gcd(100, 75) == 25)
assert(gcd(17, 13) == 1)
assert(gcd(0, 5) == 5)

# Recursive list reversal
fun reverseList(lst) {
    if (len(lst) <= 1) {
        return lst
    }
    return reverseList(lst[1..]) + [lst[0]]
}

assert(reverseList([1, 2, 3, 4, 5]) == [5, 4, 3, 2, 1])
assert(reverseList([1, 2]) == [2, 1])
assert(reverseList([]) == [])

# Recursive palindrome check
fun isPalindrome(s) {
    if (len(s) <= 1) {
        return true
    }
    if (s[0] != s[len(s) - 1]) {
        return false
    }
    return isPalindrome(s[1..len(s) - 1])
}

assert(isPalindrome("racecar") == true)
assert(isPalindrome("hello") == false)
assert(isPalindrome("a") == true)
assert(isPalindrome("") == true)

# Recursive flatten
fun flatten(lst) {
    var result = []
    for (item in lst) {
        if (type(item) == "LIST") {
            result = result + flatten(item)
        } else {
            result << item
        }
    }
    return result
}

assert(flatten([1, [2, 3], [4, [5, 6]]]) == [1, 2, 3, 4, 5, 6])
assert(flatten([]) == [])
assert(flatten([1, 2, 3]) == [1, 2, 3])

# Recursive deep copy
fun deepCopy(obj) {
    if (type(obj) == "LIST") {
        return [deepCopy(item) for item in obj]
    }
    if (type(obj) == "MAP") {
        var result = {}
        for ([k, v] in obj) {
            result[k] = deepCopy(v)
        }
        return result
    }
    return obj
}

val original = {a: [1, 2], b: {c: 3}}
val copy = deepCopy(original)
assert(copy == original)
assert(type(copy) == "MAP")
