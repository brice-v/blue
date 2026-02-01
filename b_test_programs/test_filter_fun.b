fun isEven(x) {
    x % 2 == 0
}

val abc = [1,2,3,4,5];
val expected_abc = [2,4];

assert(abc.filter(isEven) == expected_abc);

println(filter([1,2,3,4,5], isEven));
assert(filter(abc, isEven) == expected_abc)