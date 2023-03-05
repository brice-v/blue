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

### TODO: BIG_INTEGER and BIG_FLOAT are also valid so need to support them
try {
    j = (2 ** 100) + 0.5;
    println("j = #{j}, j.type() = #{j.type()}");
    j = "#{2 ** 100}";
    println("j = #{j}");
    println("j.type() = #{j.type()}");
    assert(j.is_valid_json());
    assert(false, "unreachable");
} catch (e) {
    println("e = #{e}");
    assert("unreachable" notin e);
}
###