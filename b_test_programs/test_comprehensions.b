
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
    return false;
}

println("Elements is: #{elems}")

if (alsoElems != alsoExpected) {
    return false;
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
    return false;
}
# You only append to the List (x) when both the if condition and for condition are true
var x = [x for (x in 1..4) if (x % 2 == 0)];
println("x is #{x}");
var expectedx = [2,4];

if (x != expectedx) {
    return false;
} else {
    return true;
}


var y = {n: n*2 for (n in 1..4)};
var expectedy = {1: 2, 2: 4, 3: 6, 4: 8};

if (y != expectedy) {
    return false;
}

var asdfasdf = 10;
var newasdf = {n: n**2 for (n in 1..x) if (n % 2 == 0)};
var expectednewasdf = {0: 0, 2: 4, 4: 16, 6: 36, 8: 64};
if (newasdf != expectednewasdf) {
    return false;
}


var setCompAbc = {aaaa for (aaaa in 1..10) if (aaaa % 2 == 0)};
var expectedSetAbc = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10};
if (setCompAbc != expectedSetAbc) {
    return false;
}

return true;