var x = {1,2,3,4,5};
var y = {1,2,3};
var z = {[1], "Abd", fun() { "Hello" }};

var union_result = (x | y ) == {1,2,3,4,5};
println("union_result = #{union_result}")
if (not union_result) {
    assert(false, "not union_result failed");
}
var symmetric_difference = (x ^ y) == {4,5};
println("symmetric_difference = #{symmetric_difference} which = #{(x ^ y)}")
if (not symmetric_difference) {
    assert(false, "not symmetric_difference failed");
}
var in_result = (( 1 in x) and (1 in y)) == true;
println("1 notin x? = #{1 notin x} (should be false)");
println("in_result = #{in_result}")
if (not in_result) {
    assert(false, "not in_result failed");
}
var in_result_one = ([1] in z) == true;
println("in_result_one = #{in_result_one}")
if (not in_result_one) {
    assert(false, "not in_result_one failed");
}
var in_result_two = ("Abd" in z) == true;
println("in_result_two = #{in_result_two}")
if (not in_result_two) {
    assert(false, "not in_result_two failed");
}
var in_result_three = (fun() { "Hello" } in z) == true;
println("in_result_three = #{in_result_three}")
if (not in_result_three) {
    assert(false, "not in_result_three failed");
}
var in_result_four = ("Some" in z) != true;
println("in_result_four = #{in_result_four}")
if (not in_result_four) {
    assert(false, "not in_result_four failed");
}
var notin_result = ((7 notin x) and (9 notin y)) == true;
println("notin_result = #{notin_result}")
if (not notin_result) {
    assert(false, "not notin_result failed");
}
var intersect_result = ( x & y) == {1,2,3};
println("intersect_result = #{intersect_result}")
if (not intersect_result) {
    assert(false, "not intersect_result failed");
}

var is_subset_result = (y <= x) == true;
println("is_subset_result = #{is_subset_result}");
if (not is_subset_result) {
    assert(false, "not is_subset_result failed");
}
var is_superset_result = (x >= y) == true;
println("is_superset_result = #{is_superset_result}");
if (not is_superset_result) {
    assert(false, "not is_superset_result failed");
}
var difference_result = (x - y) == {4,5};
println("difference_result = #{difference_result} (#{x-y})");
if (not difference_result) {
    assert(false, "not difference_result failed");
}

println("x is still #{x}");
assert(x == {1,2,3,4,5});
