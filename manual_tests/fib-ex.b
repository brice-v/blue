fun fib(n) {
    if n < 2 {
        return n;
    }

    return fib(n-1) + fib(n-2);
}

#fib(28);
assert(fib(28) == 317811)