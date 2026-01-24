val json_strs = [
  # 1. Empty object
  "{}",

  # 2. Empty array
  "[]",

  # 3. Simple flat object with all primitive types
  '{"string":"hello","number":123,"float":-45.67,"boolTrue":true,"boolFalse":false,"nullValue":null}',

  # 4. Object with escaped characters in a string
  '{"escaped":"Line1\\nLine2\\tTabbed\\\\Backslash\\"Quote\\u263A"}',

  # 5. Array of homogeneous primitives
  "[1,2,3,4,5]",
  "[true,false,true]",
  "[null,null,null]",

  # 6. Array of heterogeneous values
  '["text",42,{"inner":"obj"},[1,2],false,null]',

  # 7. Nested objects (depth 3)
  '{"level1":{"level2":{"level3":"deep value"}}}',

  # 8. Nested arrays (depth 4)
  '[[[["deep"]]]]',

  # 9. Mixed nesting   object containing arrays and vice versa
  '{"matrix":[[1,2],[3,4]],"listOfObjs":[{"a":1},{"b":2}],"objWithArray":{"arr":[true,false]}}',

  # 10. Numbers in every valid form
  #"[0, -0, 123, -456, 7.89, -0.12, 1e10, -2E-3, 3.14e+2]",

  # 11. Very large integer (beyond 53 bit safe range)   still legal JSON
  '{"bigInt":9223372036854775807}',

  # 12. Zero length string and empty containers
  '""',
  "[]",
  "{}",

  # 13. Unicode characters, including surrogate pairs
  '{"emoji":"\\uD83D\\uDE00","cjk":"\\u4E2D\\u6587"}',

  # 14. Strings with all escape sequences
  '"\\b\\f\\n\\r\\t\\\\\\/\\""', # backspace, formfeed, newline, carriage return, tab, backslash, slash, quote

  # 15. Object with duplicate keys   valid JSON but parsers usually keep the last one
  '{"dup":"first","dup":"second"}',

  # 16. Array containing an empty object and an empty array
  "[{},[]]",

  # 17. Complex real world example (a tiny configuration)
  '{"version":1,"name":"Example","enabled":true,"threshold":0.75,"tags":["alpha","beta"],"metadata":{"created":"2023-08-01T12:34:56Z","owner":null}}',

  # 18. Trailing whitespace (legal)
  '   {"key":"value"}   ',

  # 19. Comment like content inside a string (JSON has no comments, but they can appear as data)
  '{"note":"# not a comment #"}',
  # 20. Deeply nested mixed structure (depth ~10)   tests stack handling
  '{"a":{"b":[{"c":{"d":[{"e":null}]}}]}}',
  "12676506002282294014967032053760000000000.512676506002282294014967032053760000000000",
  "false",
  "123",
  "12676506002282294014967032053760000000000",
  "null",
  "0.2123"
];

for (var i = 0; i < json_strs.len(); i+= 1) {
    println(from_json(json_strs[i]));
}

var j = "1";
var expected = 1;
assert(j.is_valid_json());
assert(j.from_json() == expected);

j = "null";
expected = null;
assert(j.is_valid_json());
assert(j.from_json() == expected);

j = "0.213";
expected = 0.213;
assert(j.is_valid_json());
assert(j.from_json() == expected);

j = "false";
expected = false;
assert(j.is_valid_json());
assert(j.from_json() == expected);

j = """[1,3,0.123,false,{"name": "b"}]""";
expected = [1,3,0.123,false,{name: 'b'}];
assert(j.is_valid_json());
assert(j.from_json() == expected);

j = "12676506002282294014967032053760000000000.512676506002282294014967032053760000000000";
expected = 12676506002282294014967032053760000000000.512676506002282294014967032053760000000000;
assert(j.is_valid_json());
assert(j.from_json() == expected);

j = "12676506002282294014967032053760000000000";
expected = 12676506002282294014967032053760000000000;
assert(j.is_valid_json());
assert(j.from_json() == expected);

try {
    j = "{name: 123}";
    assert(j.is_valid_json());
    assert(false, "unreachable");
} catch (e) {
    println("e = #{e}");
    assert("unreachable" notin e);
}
