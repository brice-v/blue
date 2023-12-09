val strs = ["quick", "brown", "fox", "jumps"];
val strs_sorted = ['brown', 'fox', 'jumps', 'quick'];
val strs_sorted_rev = ['quick', 'jumps', 'fox', 'brown'];
val ints = [56, 19, 78, 67, 14, 25];
val ints_sorted = [14, 19, 25, 56, 67, 78];
val ints_sorted_rev = [78, 67, 56, 25, 19, 14];
val floats = [176.8, 19.5, 20.8, 57.4];
val floats_sorted = [19.5, 20.8, 57.4, 176.8];
val floats_sorted_rev = [176.8, 57.4, 20.8, 19.5];

assert(strs.sort() == strs_sorted);
assert(ints.sort() == ints_sorted);
assert(floats.sort() == floats_sorted);
assert(strs.sort(reverse=true) == strs_sorted_rev);
assert(ints.sort(reverse=true) == ints_sorted_rev);
assert(floats.sort(reverse=true) == floats_sorted_rev);


val mixed_list = ['quick', 1];
try {
    mixed_list.sort();
    assert(false);
} catch (e) {
    assert(e == "EvaluatorError: `sort` error: all elements in list must be STRING, INTEGER, or FLOAT");
}
assert(true);