
var elems = [];
var alsoElems = [];

for (x in 1..10) {
    elems = append(elems, x)
}

for (x in 1..<10) {
    println(x)
    alsoElems = append(alsoElems, x)
}
var expectedElems = [1,2,3,4,5,6,7,8,9,10];
var alsoExpected = [1,2,3,4,5,6,7,8,9];

if (elems != expectedElems) {
    false
}

println("Elements is: #{elems}")

if (alsoElems != alsoExpected) {
    false
}

println("Also Elems is: #{alsoElems}")

var count = 0;
for (i in 1..5) {
    for (x in 1..2) {
        count += 1;
    }
}
println("Count is #{count}");
if (count != 10) {
    false
}
# You only append to the List (x) when both the if condition and for condition are true
var x = [x for (x in 1..4) if (x % 2 == 0)];
println("x is #{x}");
var expectedx = [2,4];

if (x != expectedx) {
    false
} else {
    true
}

###
var y = {n: n*2 for n in 1..4}
var expectedy = {1: 2, 2: 4, 3: 6, 4: 8}

if (y != expectedy) {
    false
}
###

#true