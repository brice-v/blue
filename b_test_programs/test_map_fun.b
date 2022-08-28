fun addOne(x) {
    x+1
}

val abc = [1,2,3,4,5];
val expected_abc = [2,3,4,5,6];
if (abc.map(addOne) != expected_abc) {
    return false;
}
if (map(abc, addOne) != expected_abc) {
    return false;
}
println(map([1,2,3,4,5], addOne));

return true;