var xyz = 1;
for (i in {1, 2, 3, 4, 5}) {
    assert(xyz == i);
    xyz += 1;
}

for ([a, b] in {'a', 'b', 'c'}) {
    println("a=#{a}, b=#{b}");
    if (a == 0) {
        assert(b == 'a');
    }
    if (a == 1) {
        assert(b == 'b');
    }
    if (a == 2) {
        assert(b == 'c');
    }
}

assert(true);