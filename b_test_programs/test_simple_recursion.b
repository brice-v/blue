fun fib(n) {
    var j = 10;
    if n < 2 {
        return n;
    }

    return (fib(n-1) + fib(n-2)) * j;
}

println(fib(2));
assert(fib(2) == 10);