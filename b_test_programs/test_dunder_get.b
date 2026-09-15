## `__get` and `__len` let a map act as a custom iterable so `for in` uses it.

fun make_iterable(values) {
    var this = {};
    this._values = values;
    this.__len = fun() {
        return len(this._values);
    }
    this.__get = fun(index, with_index=false) {
        if with_index {
            return [index, this._values[index]];
        }
        return this._values[index];
    }
    return this;
}

val iterable = make_iterable([10, 20, 30]);
assert(len(iterable) == 3);

var out = [];
for v in iterable {
    out << v;
}
assert(out == [10, 20, 30]);

var pairs = [];
for [i, v] in make_iterable(["a", "b"]) {
    pairs << [i, v];
}
assert(pairs == [[0, "a"], [1, "b"]]);

# A plain map still uses the builtin `_get_` and `len`.
val plain = {first: 1, second: 2};
assert(len(plain) == 2);
var plain_out = [];
for v in plain {
    plain_out << v;
}
assert(plain_out == [1, 2]);
