fun totalUp(total, amount) {
    total + amount
}

val abc = [1,2,3,4,5];
val expected_abc = 15;
if (abc.reduce(totalUp) != expected_abc) {
    false
}
if (reduce(abc, totalUp) != expected_abc) {
    false
}
println(reduce([1,2,3,4,5], totalUp));

true