#VM IGNORE
# Sorting with null values and null sort keys.
val ints = [3, null, 1, 2];
assert(ints.sort() == [1, 2, 3, null]);

var ints_rev = [3, null, 1, 2];
assert(ints_rev.sorted(reverse=true) == null);
assert(ints_rev == [null, 3, 2, 1]);

val users = [{name: "a", age: 2}, {name: "b", age: null}, {name: "c", age: 1}];
val by_age = users.sort(key=|u| => u.age);
assert(by_age == [{name: "c", age: 1}, {name: "a", age: 2}, {name: "b", age: null}]);

val by_age_rev = users.sort(key=|u| => u.age, reverse=true);
assert(by_age_rev == [{name: "b", age: null}, {name: "a", age: 2}, {name: "c", age: 1}]);

var users_mut = new(users);
assert(users_mut.sorted(key=[|u| => u.age, |u| => u.name]) == null);
assert(users_mut == [{name: "c", age: 1}, {name: "a", age: 2}, {name: "b", age: null}]);
