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
