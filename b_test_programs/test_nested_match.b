# Match with nested patterns
val nested = {data: [1, 2, {x: 10}]}
val nestedResult = match (nested) {
    {data: [_, _, {x: 10}]} => { "matched nested" },
    _ => { "no match" }
}
assert(nestedResult == "matched nested")

# Nested match should not match wrong values
val wrongNested = {data: [1, 2, {x: 20}]}
val wrongResult = match (wrongNested) {
    {data: [_, _, {x: 10}]} => { "matched wrong" },
    _ => { "no match - correct" }
}
assert(wrongResult == "no match - correct")

# Nested match with all wildcards
val anyNested = {data: [1, 2, 3]}
val anyResult = match (anyNested) {
    {data: [_, _, _]} => { "matched any" },
    _ => { "no match" }
}
assert(anyResult == "matched any")

# List with nested maps and wildcards
val complex = {users: [{name: "Alice", age: 30}, {name: "Bob", age: 25}]}
val complexResult = match (complex) {
    {users: [{name: "Alice", age: _}, _]} => { "found Alice" },
    _ => { "no match" }
}
assert(complexResult == "found Alice")

# --- Set pattern matching ---

# Set with literal values
val s = {1, 2, 3}
val setResult1 = match (s) {
    {1, 2, 3} => { "matched set" },
    _ => { "no match" }
}
assert(setResult1 == "matched set")

# Set should not match when elements differ
val s2 = {1, 2, 4}
val setResult2 = match (s2) {
    {1, 2, 3} => { "wrong" },
    _ => { "correct - no match" }
}
assert(setResult2 == "correct - no match")

# Set with wildcard
val setResult3 = match (s) {
    {1, _, 3} => { "matched with wildcard" },
    _ => { "no match" }
}
assert(setResult3 == "matched with wildcard")

# Set with only wildcards matches any set
val setResult4 = match (s) {
    {_, _, _} => { "matched any set" },
    _ => { "no match" }
}
assert(setResult4 == "matched any set")

# Set with strings
val ss = {"a", "b", "c"}
val setResult5 = match (ss) {
    {"a", _, "c"} => { "matched string set" },
    _ => { "no match" }
}
assert(setResult5 == "matched string set")

# Extra elements in value set are fine (subset match)
val ss2 = {"a", "b", "c", "d"}
val setResult6 = match (ss2) {
    {"a", "b"} => { "subset matched" },
    _ => { "no match" }
}
assert(setResult6 == "subset matched")

# Nested set inside map
val mapWithSet = {tags: {"x", "y", "z"}}
val setResult7 = match (mapWithSet) {
    {tags: {"x", _, _}} => { "found in nested set" },
    _ => { "no match" }
}
assert(setResult7 == "found in nested set")

# Nested sets inside list
val listOfSets = [{"a", "b"}, {"c", "d"}]
val setResult8 = match (listOfSets) {
    [{_, "b"}, {"c", _}] => { "matched list of sets" },
    _ => { "no match" }
}
assert(setResult8 == "matched list of sets")

# Mixed nested: map with list containing set with nested map
val mixed = {items: [1, {x: {"inner", "values"}}]}
val setResult9 = match (mixed) {
    {items: [_, {x: {_, "values"}}]} => { "deeply nested match" },
    _ => { "no match" }
}
assert(setResult9 == "deeply nested match")

# --- Struct pattern matching ---

# Basic struct match
val st = @{one: 1, hello: "world"}
val structResult1 = match (st) {
    @{one: 1, hello: "world"} => { "matched struct" },
    _ => { "no match" }
}
assert(structResult1 == "matched struct")

# Struct match should fail on wrong values
val st2 = @{one: 2, hello: "world"}
val structResult2 = match (st2) {
    @{one: 1, hello: "world"} => { "wrong" },
    _ => { "correct - no match" }
}
assert(structResult2 == "correct - no match")

# Struct match with wildcard
val structResult3 = match (st) {
    @{one: _, hello: "world"} => { "matched with wildcard" },
    _ => { "no match" }
}
assert(structResult3 == "matched with wildcard")

# Struct match with all wildcards
val structResult4 = match (st) {
    @{one: _, hello: _} => { "matched any struct" },
    _ => { "no match" }
}
assert(structResult4 == "matched any struct")

# Struct with nested map
val stWithMap = @{data: {x: 10, y: 20}, label: "point"}
val structResult5 = match (stWithMap) {
    @{data: {x: 10, y: _}, label: _} => { "matched nested map in struct" },
    _ => { "no match" }
}
assert(structResult5 == "matched nested map in struct")

# Struct with nested list
val stWithList = @{items: [1, 2, 3], name: "test"}
val structResult6 = match (stWithList) {
    @{items: [_, 2, _], name: _} => { "matched nested list in struct" },
    _ => { "no match" }
}
assert(structResult6 == "matched nested list in struct")

# Struct with nested set
val stWithSet = @{tags: {"a", "b", "c"}, id: 42}
val structResult7 = match (stWithSet) {
    @{tags: {"a", _, _}, id: _} => { "matched nested set in struct" },
    _ => { "no match" }
}
assert(structResult7 == "matched nested set in struct")

# Nested structs
val inner = @{x: 10, y: 20}
val outer = @{point: inner, desc: "nested"}
val structResult8 = match (outer) {
    @{point: @{x: 10, y: _}, desc: _} => { "matched nested struct" },
    _ => { "no match" }
}
assert(structResult8 == "matched nested struct")

# List of structs
val structList = [@{name: "Alice", age: 30}, @{name: "Bob", age: 25}]
val structResult9 = match (structList) {
    [@{name: "Alice", age: _}, _] => { "found Alice in list" },
    _ => { "no match" }
}
assert(structResult9 == "found Alice in list")

# Map with struct values
val structMap = {first: @{a: 1}, second: @{a: 2}}
val structResult10 = match (structMap) {
    {first: @{a: 1}, second: _} => { "matched struct value in map" },
    _ => { "no match" }
}
assert(structResult10 == "matched struct value in map")

# Set of structs
val structSet = {@{x: 1}, @{x: 2}, @{x: 3}}
val structResult11 = match (structSet) {
    {@{x: 1}, @{x: 2}, _} => { "matched struct in set" },
    _ => { "no match" }
}
assert(structResult11 == "matched struct in set")
