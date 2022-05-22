# This tests that we are populating the remaining areas of the list with null (that arent assigned)
var x = [];

x[10] = "HELLO";

for ([i, e] in x) {
    println("i=#{i}, e=#{e}");
}

var y = [];

for (i in 1..<10) {
    y[i] = null;
}

y[10] = "HELLO";

if (x != y) {
    false
}

true