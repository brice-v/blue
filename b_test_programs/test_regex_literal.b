var x = r/abc[\t|\s]/;

var content = "abc\t";
println(content.matches(x));
assert(content.matches(x));
x = r/abc[\/|\s]/;
content = "abc/";
println(content.matches(x));
assert(content.matches(x));
content = "abc ";
println(content.matches(x));
assert(content.matches(x));