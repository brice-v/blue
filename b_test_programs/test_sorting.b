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
assert(strs.sorted() == null);
assert(strs == strs_sorted);
assert(ints.sort() == ints_sorted);
assert(ints.sorted() == null);
assert(ints == ints_sorted);
assert(floats.sort() == floats_sorted);
assert(floats.sorted() == null);
assert(floats == floats_sorted);
assert(strs.sort(reverse=true) == strs_sorted_rev);
assert(strs.sorted(reverse=true) == null);
assert(strs == strs_sorted_rev);
assert(ints.sort(reverse=true) == ints_sorted_rev);
assert(ints.sorted(reverse=true) == null);
assert(ints == ints_sorted_rev);
assert(floats.sort(reverse=true) == floats_sorted_rev);
assert(floats.sorted(reverse=true) == null);
assert(floats == floats_sorted_rev);


val mixed_list = ['quick', 1];
try {
    mixed_list.sort();
    assert(false);
} catch (e) {
    assert(e == "`sort` error: all elements in list must be STRING, INTEGER, or FLOAT");
}


val users = [{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steve", age: 28}];
val users1 = new(users);
var users2 = new(users);
val users_sorted = users.sort(key=fun(user) { return user.age });
val users_sorted_rev = users.sort(key=|user| => user.age, reverse=true);
val users_sorted_expected = [{name: 'Amanda', age: 16}, {name: 'Rajeev', age: 28}, {name: 'Steve', age: 28}, {name: 'Monica', age: 31}, {name: 'John', age: 56}];
val users_sorted_expected_rev = [{name: 'John', age: 56}, {name: 'Monica', age: 31}, {name: 'Rajeev', age: 28}, {name: 'Steve', age: 28}, {name: 'Amanda', age: 16}];
println("users        = #{users}")
println("users_sorted = #{users_sorted}")
println([{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steve", age: 28}] == [{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steve", age: 28}]);
assert([{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steve", age: 28}] == [{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steve", age: 28}]);
assert(users == [{name: "Rajeev", age: 28}, {name: "Monica", age: 31}, {name: "John", age: 56}, {name: "Amanda", age: 16}, {name: "Steve", age: 28}]);
assert(users_sorted == users_sorted_expected);
assert(users_sorted == [{name: 'Amanda', age: 16}, {name: 'Rajeev', age: 28}, {name: 'Steve', age: 28}, {name: 'Monica', age: 31}, {name: 'John', age: 56}]);
assert(users1.sorted(key=fun(user) { return user.age }) == null);
assert(users1 == users_sorted_expected);
println("users_sorted_rev = #{users_sorted_rev}")
assert(users_sorted_rev == users_sorted_expected_rev);
assert(users_sorted_rev == [{name: 'John', age: 56}, {name: 'Monica', age: 31}, {name: 'Rajeev', age: 28}, {name: 'Steve', age: 28}, {name: 'Amanda', age: 16}]);
assert(users2.sorted(key=|user| => user.age, reverse=true) == null);
assert(users2 == users_sorted_expected_rev);


val users_sorted_2keys = users.sort(key=[|user| => user.age, |user| => len(user.name)]);
var users3 = new(users);
println("users_sorted_2keys = #{users_sorted_2keys}")
val users_sorted_2keys_expected = [{name: 'Amanda', age: 16}, {name: 'Steve', age: 28}, {name: 'Rajeev', age: 28}, {name: 'Monica', age: 31}, {name: 'John', age: 56}];
val users_sorted_2keys_rev = users.sort(key=[|user| => user.age, |user| => user.name], reverse=true);
var users4 = new(users);
println("users_sorted_2keys_rev = #{users_sorted_2keys_rev}")
val users_sorted_2keys_expected_rev = [{name: 'John', age: 56}, {name: 'Monica', age: 31}, {name: 'Steve', age: 28}, {name: 'Rajeev', age: 28}, {name: 'Amanda', age: 16}];
assert(users_sorted_2keys == users_sorted_2keys_expected);
assert(users3.sorted(key=[|user| => user.age, |user| => len(user.name)]) == null);
assert(users3 == users_sorted_2keys_expected);
assert(users_sorted_2keys_rev == users_sorted_2keys_expected_rev);
assert(users4.sorted(key=[|user| => user.age, |user| => user.name], reverse=true) == null);
assert(users4 == users_sorted_2keys_expected_rev);


val set_ex = [{"quick", "brown", 1}, {"fox", "jumps", 2}];
println("set_ex = #{set_ex}")
val set_ex_sorted = set_ex.sort(key=|e| => e[0]);
println("set_ex_sorted = #{set_ex_sorted}")
assert(set_ex_sorted == [{"fox", "jumps", 2}, {"quick", "brown", 1}]);
assert(set_ex.sorted(key=|e| => e[0]) == null);
assert(set_ex == set_ex_sorted);
val list_ex = [["quick", "brown", 2], ["fox", "jumps", 1]];
println("list_ex = #{list_ex}")
val list_ex_sorted = list_ex.sort(key=|e| => e[2]);
println("list_ex_sorted = #{list_ex_sorted}")
assert(list_ex_sorted == [["fox", "jumps", 1], ["quick", "brown", 2]])
assert(list_ex.sorted(key=|e| => e[2]) == null);
assert(list_ex == list_ex_sorted);
val obj_int_key = [{28: "Rajeev"}, {31: "Monica"}, {56: "John"}, {16: "Amanda"}, {28: "Steve"}];
println("obj_int_key = #{obj_int_key}")
val obj_int_key_sorted = obj_int_key.sort(key=|e| => e.keys()[0]);
println("obj_int_key_sorted = #{obj_int_key_sorted}")
assert(obj_int_key_sorted == [{16: "Amanda"},{28: "Rajeev"},{28: "Steve"},{31: "Monica"},{56: "John"}]);
assert(obj_int_key.sorted(key=|e| => e.keys()[0]) == null)
assert(obj_int_key == obj_int_key_sorted)
val obj_float_key = [{19.5: "Rajeev"}, {176.8: "Monica"}, {20.8: "John"}, {57.4: "Amanda"}];
println("obj_float_key = #{obj_float_key}")
val obj_float_key_sorted = obj_float_key.sort(key=|e| => e.keys()[0]);
println("obj_float_key_sorted = #{obj_float_key_sorted}")
assert(obj_float_key_sorted == [{19.5: "Rajeev"},{20.8: "John"},{57.4: "Amanda"},{176.8: "Monica"}]);
assert(obj_float_key.sorted(key=|e| => e.keys()[0]) == null);
assert(obj_float_key == obj_float_key_sorted);


assert(true);