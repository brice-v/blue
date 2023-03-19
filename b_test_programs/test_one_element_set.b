val x = 123;
val y = {x};
val z = {123};

assert(type(y) == Type.SET);
assert(type(z) == Type.SET);

assert(x in y);
assert(x in z);
assert(y[0] == 123);
assert(z[0] == 123);
assert(4 notin y);
assert(4 notin z);