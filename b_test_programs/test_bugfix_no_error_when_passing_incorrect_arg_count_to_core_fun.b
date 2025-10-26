val left = [1,1,3,1,1];
val right = [1,1,5,'a','b'];

try {
    zip(left, right);
} catch (e) {
    assert(e == "function called with too many arguments");
}

assert(true);