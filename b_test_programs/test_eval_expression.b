#VM IGNORE
val x = 1;
val y = 1;

val z = eval("x + #{y}");
assert(z == 2);
eval("println(#{z})");

try {
    eval(1);
    error("`eval` should only be possible with strings");
} catch (e) {
    assert(e == "value after `eval` must be STRING. got INTEGER");
}

eval("println('Hello World!')");

assert(true);