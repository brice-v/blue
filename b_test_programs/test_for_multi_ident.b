var abca = [1, 2, 3];
var abcb = [];

for ([a, b] in abca) {
    println("a=#{a}, b=#{b}");
    abcb[a] = b;
}

assert(abca == abcb);

var x = {some: "world", another: "thing"};
var z = {};

for ([a, b] in x) {
    println("a=#{a}, b=#{b}");
    z[a] = b;
}

assert(z == x);

var xxx = "Hello World!";
var zzz = xxx.split("");

var xyz = [];

for ([a, b] in xxx) {
    println("a=#{a}, b=#{b}");
    xyz[a] = b;
}

println("Here!");
println("xyz = #{xyz}, zzz = #{zzz}, (xyz != zzz)=#{xyz != zzz}");
assert(xyz == zzz);
println("HERE 2");
assert(true);