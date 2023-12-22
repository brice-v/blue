var x = r/abc[\t|\s]/;
var xx = re("abc[\\t|\\s]");
assert(xx.type() == Type.REGEX);

var content = "abc\t";
println(content.matches(x));
assert(content.matches(x));
assert(content.matches(xx));
x = r/abc[\/|\s]/;
xx = re("abc[\\/|\\s]");
content = "abc/";
println(content.matches(x));
assert(content.matches(x));
assert(content.matches(xx));
content = "abc ";
println(content.matches(x));
assert(content.matches(x));
assert(content.matches(xx));