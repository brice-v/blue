# Calling a function inside of a match pattern (alongside wildcards) should work.
# Guards that function bodies within match conditions are isolated from '_' wildcard handling.
val r = match ([5, 2]) {
    [(fun(x) { return x })(5), _] => { "matched" },
    _ => { "no match" }
}
println(r)
assert(r == "matched")

val r2 = match ([4, 2]) {
    [(fun(x) { return x })(5), _] => { "wrong" },
    _ => { "no match - correct" }
}
assert(r2 == "no match - correct")

# Lambda inside of a pattern as well
val r3 = match ([1, 9]) {
    [|x| => { return x }(1), _] => { "matched lambda" },
    _ => { "no match" }
}
println(r3)
assert(r3 == "matched lambda")

# Wildcards still work normally in patterns after functions were compiled in earlier arms
val r4 = match ({a: [1, 2], b: 3}) {
    {a: [_, _], b: _} => { "wildcards still fine" },
    _ => { "no match" }
}
println(r4)
assert(r4 == "wildcards still fine")
