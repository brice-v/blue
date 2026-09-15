## Multi-condition match arms evaluate conditions left to right and stop at the
## first true one; all of them share a single copy of the arm's body.

var calls = [];
fun guard(name, result) {
    calls << name;
    return result;
}

calls = [];
var r = match {
    guard("first", true), guard("second", true) => { "matched" },
    _ => { "else" },
};
assert(r == "matched");
assert(calls == ["first"]);

calls = [];
r = match {
    guard("first", false), guard("second", true) => { "matched" },
    _ => { "else" },
};
assert(r == "matched");
assert(calls == ["first", "second"]);

calls = [];
r = match {
    guard("first", false), guard("second", false) => { "matched" },
    _ => { "else" },
};
assert(r == "else");
assert(calls == ["first", "second"]);

# A default condition matches without testing the conditions after it.
calls = [];
r = match {
    _, guard("never", true) => { "default" },
};
assert(r == "default");
assert(calls == []);

println("match short circuit pass")
