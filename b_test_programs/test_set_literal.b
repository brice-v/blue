var x = {1,2,3,4,5};
var y = {1,2,3};
var z = {[1], "Abd", fun() { "Hello" }};

var union_result = (x | y ) == {1,2,3,4,5};
println("union_result = #{union_result}")
if (not union_result) {
    return false;
}
var symmetric_difference = (x ^ y) == {4,5};
println("symmetric_difference = #{symmetric_difference} which = #{(x ^ y)}")
if (not symmetric_difference) {
    return false;
}
var in_result = (( 1 in x) and (1 in y)) == true;
println("1 notin x? = #{1 notin x} (should be false)");
println("in_result = #{in_result}")
if (not in_result) {
    return false;
}
var in_result_one = ([1] in z) == true;
println("in_result_one = #{in_result_one}")
if (not in_result_one) {
    return false;
}
var in_result_two = ("Abd" in z) == true;
println("in_result_two = #{in_result_two}")
if (not in_result_two) {
    return false;
}
var in_result_three = (fun() { "Hello" } in z) == true;
println("in_result_three = #{in_result_three}")
if (not in_result_three) {
    return false;
}
var in_result_four = ("Some" in z) != true;
println("in_result_four = #{in_result_four}")
if (not in_result_four) {
    return false;
}
var notin_result = ((7 notin x) and (9 notin y)) == true;
println("notin_result = #{notin_result}")
if (not notin_result) {
    return false;
}
var intersect_result = ( x & y) == {1,2,3};
println("intersect_result = #{intersect_result}")
if (not intersect_result) {
    return false;
}


var is_subset_result = (y <= x) == true;
if (not is_subset_result) {
    return false;
}
var is_superset_result = (x >= y) == true;
if (not is_superset_result) {
    return false;
}
var difference_result = (x - y) == {4,5};
if (not difference_result) {
    return false;
}

println("x is still #{x}");
if (x != {1,2,3,4,5}) {
    return false;
} else {
    return true;
}



