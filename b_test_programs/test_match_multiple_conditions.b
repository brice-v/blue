var result = "";

result += "T1: ";
match {
    true => { println("T1 pass") },
    _ => { println("T1 FAIL"); assert(false) },
}

result = "";
result += match 2 {
    1 => { "wrong" },
    2, 3 => { "matched_second" },
    _ => { "no_match" },
};
assert(result == "matched_second")

result = "";
result += match 7 {
    1, 2 => { "one_or_two" },
    3, 4, 5 => { "three_to_five" },
    _ => { "default_matched" },
};
assert(result == "default_matched")

result = "";
result += match "hello" {
    "world", "foo" => { "nope" },
    "bar", "baz", "hello" => { "greeted" },
    _ => { "else" },
};
assert(result == "greeted")

fun test_both_guards_true() {
    var x = 5;
    var y = 3;
    match {
        x > 0, y < 10 => { return "both_true" },
        _ => { return "else" },
    }
}
var r1 = test_both_guards_true();
assert(r1 == "both_true")

fun test_second_guard_false_but_first_true() {
    var x = 5;
    var y = 20;
    match {
        x > 0, y < 10 => { return "both_true" },
        _ => { return "else" },
    }
}
var r2 = test_second_guard_false_but_first_true();
assert(r2 == "both_true")

fun test_catch_all() {
    var x = 5;
    var y = 3;
    match {
        x < 0, y < 0 => { return "negative" },
        _, _ => { return "catch_all" },
        x > 0, y > 0 => { return "should_not_reach" },
    }
}
var r3 = test_catch_all();
assert(r3 == "catch_all")

result = "";
result += match "a" {
    "a" => { "single" },
    _ => { "other" },
};
assert(result == "single")
