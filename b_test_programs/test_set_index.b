val x = {y for (y in 1..10)};
val z = {1, 2, 3, 4, 5}

val x1 = x[0];
# This is not possible with current parser (not planning on supporting it)
#val x2 = x.x.len();
val x2 = x[len(x)];
val z2 = z.0;
val z3 = z[z.len()]

if (x1 != 1) {
    false
}
if (x2 != 10) {
    false
}
if (z2 != 1) {
    false
}
if (z3 != 5) {
    false
}

if ({y for (y in 1..10)}.0 != 1) {
    false
}

true