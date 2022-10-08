val identKeyWithQuote = """Hello"World!""";
val identForValue = 1241092401924;

val xyz = {
    "one": 1,
    "two": 2.3134,
    identKeyWithQuote: {
        three: 142214124,
        "four": 2842.124,
        five: true
    },
    "after_hello_world": null,
    six: [1, 2, 3, null, 2.123123, {seven: "bced"}, ["a", "b", "c"]],
    last_elem: identForValue
};

# Looks like it stays the same within the same evaluation (but each run is different)

println(xyz.to_json());
val actual_json = xyz.to_json();

val expected_json = """{"one":1,"two":2.313400,"Hello\"World!":{"three":142214124,"four":2842.124000,"five":true},"after_hello_world":null,"six":[1,2,3,null,2.123123,{"seven":"bced"},["a","b","c"]],"last_elem":1241092401924}""";

if (actual_json != expected_json) {
    false
} else {
    true
}

val expected_json_back_to_map = expected_json.json_to_map();
println("expected_json_back_to_map = #{expected_json_back_to_map}");

if (expected_json_back_to_map != xyz) {
    false
} else {
    true
}