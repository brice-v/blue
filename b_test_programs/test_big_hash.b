for (i in 1..2) {
    var z = "def";
}
for (i in 1..10) {
    var z = "def";
}


val xyz = 100.210389218302108380123 * 2.028140812;

val z = set([xyz]);

assert(true);

if (xyz in z) {
    assert(true);
} else {
    error("xyz not in z");
}


if (xyz notin z) {
    error("xyz should be in z");
} else {
    assert(true);
}

assert(true);