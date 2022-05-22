var abca = [1, 2, 3];
var abcb = [];

for ([a, b] in abca) {
    println("a=#{a}, b=#{b}");
    abcb[a] = b;
}

if (abca != abcb) {
    false
}

var x = {some: "world", another: "thing"};
var z = {};

for ([a, b] in x) {
    println("a=#{a}, b=#{b}");
    z[a] = b;
}

if (z != x) {
    false
}

var xxx = "Hello World!";
var zzz = xxx.split("");

var xyz = [];

for ([a, b] in xxx) {
    println("a=#{a}, b=#{b}");
    xyz[a] = b;
}

if (xyz != zzz) {
    false
}


true
