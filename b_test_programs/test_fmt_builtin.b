var x = 1;
var y = 0x20;
var z = "HELLO";
var a = 1.234;

var test = x.fmt("%04b")
println(test);
assert(test == "0001")
test = y.fmt("%03X")
println(test);
assert(test == "020")
test = z.fmt("%s World!")
println(test);
assert(test == "HELLO World!")
test = a.fmt("%0.6f")
println(test);
assert(test == "1.234000")


var abc = {hello: 'world'};
test = abc.fmt("%v")
println(test);
assert(test == "&{map[hello:world] [hello]}");

abc[123] = 903;
test = abc.fmt("%q");
println(test);
assert(test == '"{hello: world, 123: 903}"');