var x = [1,2,3,4,5];
if (3 notin x) {
    false
}
var z = [1.0,2.0,3.0];
if (3.0 notin z) {
    false
}
var y = "Some String"
if ("Some" notin y) {
    false
}
if ("some" in y) {
    false
}

var abc = {name: "brice", key: "another"}
if ("key" in abc) {
    true
}
var someother = [abc, 123, 0x100]
if (abc notin someother) {
    false
}
if (0x100 in someother) {
    true
} else {
    false
}

true