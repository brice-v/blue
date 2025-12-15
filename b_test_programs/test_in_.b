var x = [1,2,3,4,5];
if (3 notin x) {
    assert(false)
}
var z = [1.0,2.0,3.0];
if (3.0 notin z) {
    assert(false)
}
var y = "Some String"
if ("Some" notin y) {
    assert(false)
}
if ("some" in y) {
    assert(false)
}

var abc = {name: "brice", key: "another"}
assert("key" in abc);
var someother = [abc, 123, 0x100]
if (abc notin someother) {
    assert(false)
}
if (0x100 in someother) {
    assert(true)
} else {
    assert(false)
}


val abc123 = {1,2,3,4,5};
if (1 notin abc123) {
    assert(false);
}
if (1 in abc123) {
    assert(true)
} else {
    assert(false);
}

if (100 in abc123) {
    assert(false);
}

assert(true);