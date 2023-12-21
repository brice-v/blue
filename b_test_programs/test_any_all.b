val x = 1..10;
val y = {'a': 1, 'b': 1};

assert(x.any(|e| => e == 5));
assert(y.values().all(|e| => e == 1));
