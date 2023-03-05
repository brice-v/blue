fun hello(x, y = 3+2, z = 4, a) {
    x + y + z + a
}

assert(hello(3,5) == 17)