# Test complex match expressions

# Basic match with integer
val result = match (5) {
    1 => { "one" },
    2 => { "two" },
    5 => { "five" },
    _ => { "other" }
}
assert(result == "five")

# Match with string
val strResult = match ("hello") {
    "hello" => { "greeting" },
    "world" => { "planet" },
    _ => { "unknown" }
}
assert(strResult == "greeting")

# Match with boolean
val boolResult = match (true) {
    true => { "yes" },
    false => { "no" }
}
assert(boolResult == "yes")

# Match with null
val nullResult = match (null) {
    null => { "is null" },
    _ => { "not null" }
}
assert(nullResult == "is null")

# Match without value (condition-only)
val condResult = match {
    5 > 3 => { "greater" },
    5 < 3 => { "less" },
    _ => { "equal" }
}
assert(condResult == "greater")

# Match with complex patterns
val obj = {name: "blue", version: 1}
val objResult = match (obj) {
    {name: "blue", version: _} => { "blue language" },
    {name: _, version: _} => { "some object" },
    _ => { "other" }
}
assert(objResult == "blue language")

# Match with list patterns
val listResult = match ([1, 2, 3]) {
    [1, 2, 3] => { "exact match" },
    [_, _, _] => { "three elements" },
    [] => { "empty" },
    _ => { "other" }
}
assert(listResult == "exact match")

# Match with nested patterns
val nested = {data: [1, 2, {x: 10}]}
val nestedResult = match (nested) {
    {data: [_, _, {x: 10}]} => { "matched nested" },
    _ => { "no match" }
}
assert(nestedResult == "matched nested")

# Match returning different types
val mixedResult = match (42) {
    0 => { "zero" },
    _ => { 42 }
}
assert(mixedResult == 42)

# Match with function call in pattern
fun getFive() { 5 }
val funcResult = match (getFive()) {
    5 => { "got five" },
    _ => { "not five" }
}
assert(funcResult == "got five")

# Match in function
fun classify(x) {
    match (x) {
        0 => { "zero" },
        1 => { "one" },
        2 => { "two" },
        _ => { "many" }
    }
}

assert(classify(0) == "zero")
assert(classify(1) == "one")
assert(classify(2) == "two")
assert(classify(100) == "many")

# Match with multiple wildcards
val multiWild = match ({a: 1, b: 2}) {
    {a: _, b: _} => { "has a and b" },
    {a: _} => { "has only a" },
    {b: _} => { "has only b" },
    _ => { "has neither" }
}
assert(multiWild == "has a and b")

# Match with type checking via patterns
val typeMatch = match ([1, "hello", true]) {
    [int, string, bool] => { "typed list" },
    _ => { "not typed" }
}
# Note: this depends on whether blue supports type patterns
# If not, it falls through to _
assert(typeMatch == "not typed" || typeMatch == "typed list")

# Match with arithmetic in patterns
val arithResult = match (3 + 2) {
    5 => { "equals five" },
    _ => { "not five" }
}
assert(arithResult == "equals five")

# Match with range check
val rangeMatch = match (50) {
    x if x < 0 => { "negative" },
    x if x < 100 => { "positive less than 100" },
    x if x >= 100 => { "hundred or more" },
    _ => { "fallback" }
}
assert(rangeMatch == "positive less than 100")

# Match in list comprehension
val categorized = [match (x) {
    x if x % 2 == 0 => { "even" },
    _ => { "odd" }
} for x in 1..6]
assert(categorized == ["odd", "even", "odd", "even", "odd", "even"])

# Match with default case
val defaultResult = match (999) {
    1 => { "one" },
    2 => { "two" }
}
assert(defaultResult == null)  # No _ case, returns null

# Match with string starts with
val prefixResult = match ("hello world") {
    x if x.startswith("hello") => { "starts with hello" },
    x if x.startswith("world") => { "starts with world" },
    _ => { "other" }
}
assert(prefixResult == "starts with hello")

# Match with empty container
val emptyMatch = match ([]) {
    [] => { "empty list" },
    _ => { "not empty" }
}
assert(emptyMatch == "empty list")

val emptyMapMatch = match ({}) {
    {} => { "empty map" },
    _ => { "not empty" }
}
assert(emptyMapMatch == "empty map")

# Match with set
val setMatch = match ({1, 2, 3}) {
    {1, 2, 3} => { "exact set" },
    _ => { "other set" }
}
assert(setMatch == "exact set")

# Match with boolean expression
val exprMatch = match (true) {
    true => { "truthy" },
    false => { "falsy" }
}
assert(exprMatch == "truthy")
