val x = [1,2,3,4,5,6,7,8,9,10];

val y = x[1..3];
val yy = x[1..<4];
println("y = #{y}");
println("yy = #{yy}");

val z = [2,3,4];

assert(y == z);
assert(yy == z);

# TODO: Do the same test with sets and strings
# TODO: Also need to see how this would work for setting things in a list?

val a = {1,2,3,4,5,6,7,8,9,10};

val y1 = a[1..3];
val yy1 = a[1..<4];
println("y1 = #{y1}");
println("yy1 = #{yy1}");

val z1 = {2,3,4};

assert(y1 == z1);
assert(yy1 == z1);

val aa = 'abcdefghij';

val y2 = aa[1..3];
val yy2 = aa[1..<4];
println("y2 = #{y2}");
println("yy2 = #{yy2}");

val z2 = 'bcd';

assert(y2 == z2);
assert(yy2 == z2);