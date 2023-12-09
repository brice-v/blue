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


val users = [{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steven", age: 28}];
val users_sorted = users.sort(key=fun(user) { return user.age });
val users_sorted_rev = users.sort(key=|user| => user.age, reverse=true);
val users_sorted_expected = [{name: 'Amanda', age: 16}, {name: 'Rajeev', age: 28}, {name: 'Steven', age: 28}, {name: 'Monica', age: 31}, {name: 'John', age: 56}];
val users_sorted_expected_rev = [{name: 'John', age: 56}, {name: 'Monica', age: 31}, {name: 'Rajeev', age: 28}, {name: 'Steven', age: 28}, {name: 'Amanda', age: 16}];
println("users        = #{users}")
println("users_sorted = #{users_sorted}")
println([{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steven", age: 28}] == [{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steven", age: 28}]);
assert([{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steven", age: 28}] == [{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steven", age: 28}]);
assert(users == [{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steven", age: 28}]);
assert(users_sorted == users_sorted_expected);
assert(users_sorted == [{name: 'Amanda', age: 16}, {name: 'Rajeev', age: 28}, {name: 'Steven', age: 28}, {name: 'Monica', age: 31}, {name: 'John', age: 56}]);
println("users_sorted_rev = #{users_sorted_rev}")
assert(users_sorted_rev == users_sorted_expected_rev);
assert(users_sorted_rev == [{name: 'John', age: 56}, {name: 'Monica', age: 31}, {name: 'Rajeev', age: 28}, {name: 'Steven', age: 28}, {name: 'Amanda', age: 16}]);

assert(true);