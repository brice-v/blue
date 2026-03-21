fun totalUp(total, amount) {
    if (total == null) {
        total = 0;
    }
    total + amount;
}

val abc = [1,2,3,4,5];
val expected_abc = 15;
println("abc.reduce(totalUp) = #{abc.reduce(totalUp)}");
assert(abc.reduce(totalUp) == expected_abc);
assert(reduce(abc, totalUp) == expected_abc);
println(reduce([1,2,3,4,5], totalUp));
assert(true);