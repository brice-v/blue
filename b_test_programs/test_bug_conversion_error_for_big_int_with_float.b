var x = 2 ** 100;
assert(x.type() == 'BIG_INTEGER');

var y = x + 0.5;
assert(y.type() == 'BIG_FLOAT');