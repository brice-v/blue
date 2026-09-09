import math

assert(math.exp(0.0) == 1.0);
assert(math.exp(1.0) > 2.7);
assert(math.exp(1.0) < 2.8);

assert(math.exp2(0.0) == 1.0);
assert(math.exp2(3.0) == 8.0);

assert(math.expm1(0.0) == 0.0);
assert(math.expm1(1.0) > 1.7);
assert(math.expm1(1.0) < 1.8);

assert(math.fmod(9.5, 3.0) == 0.5);

assert(math.pow(2.0, 8.0) == 256.0);
assert(math.pow(3.0, 2.0) == 9.0);
assert(math.pow(10.0, 0.0) == 1.0);

assert(math.signum(-5.0) == -1.0);
assert(math.signum(5.0) == 1.0);
assert(math.signum(0.0) == 0.0);

var seeded = math.seed(42);
assert(seeded == true);

var g1 = math.gauss(0.0, 1.0);
var g2 = math.gauss(2.5, 0.5);
println("gauss values:", g1, g2);

var items = [10, 20, 30];
var weights = [1.0, 1.0, 1.0];
var chosen = math.weighted_choice(items, weights);
println("weighted choice:", chosen);
