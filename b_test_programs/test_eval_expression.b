
val x = 1;
val y = 1;

val z = eval("x + #{y}");
if (z != 2) {
    return false;
}
eval("println(#{z})");

try {
    eval(1);
    error("`eval` should only be possible with strings");
} catch (e) {
    return e == "value after `eval` must be STRING. got INTEGER";
}

eval("println('Hello World!')");

return true;