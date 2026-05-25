fun addOne(x) {
    x+1
}

val abc = [1,2,3,4,5];
val expected_abc = [2,3,4,5,6];
assert(abc.map(addOne) == expected_abc)
println(map([1,2,3,4,5], addOne));
assert(map(abc, addOne) == expected_abc);
