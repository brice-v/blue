fun isEven(x) {
    x % 2 == 0
}

val abc = [1,2,3,4,5];
val expected_abc = [2,4];

if (abc.filter(isEven) != expected_abc) {
    return false;
}
if (filter(abc, isEven) != expected_abc) {
    return false;
}
println(filter([1,2,3,4,5], isEven));

return true;