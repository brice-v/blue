var j = [1,3,4,"abc", 0.1234, {name: "b"}];

var jj = j.to_json();

println("jj = #{jj}");

var expected = """[1,3,4,"abc",0.123400,{"name":"b"}]""";

assert(expected == jj);

j = 1;
expected = "1";
assert(j.to_json() == expected);

j = 0.1234;
expected = "0.123400";
assert(j.to_json() == expected);

j = 0x01;
expected = "1";
assert(j.to_json() == expected);

j = null;
expected = "null";
assert(j.to_json() == expected);

j = true;
expected = "true";
assert(j.to_json() == expected);

try {
    j = {1,3,4};
    j.to_json();
    assert(false, "unreachable");
} catch (e) {
    println("e = #{e}");
    assert('unreachable' notin e);
}

j = 12676506002282294014967032053760000000000.512676506002282294014967032053760000000000;
expected = "12676506002282294014967032053760000000000.51267650600228229401496703205376";
assert(j.type() == 'BIG_FLOAT');
println("j.to_json() = #{j.to_json()}");
assert(j.to_json() == expected);

j = 12676506002282294014967032053760000000000;
expected = "12676506002282294014967032053760000000000";
assert(j.type() == 'BIG_INTEGER');
assert(j.to_json() == expected);