val x = [1,2,3,4,5,6,7,8,9,10];

val y = x[1..3];
val yy = x[1..<4];
println("y = #{y}");
println("yy = #{yy}");

val z = [2,3];

assert(y == z);
assert(yy == z);

val a = {1,2,3,4,5,6,7,8,9,10};

val y1 = a[1..3];
val yy1 = a[1..<4];
println("y1 = #{y1}");
println("yy1 = #{yy1}");

val z1 = {2,3};

assert(y1 == z1);
assert(yy1 == z1);

val aa = 'abcdefghij';

val y2 = aa[1..3];
val yy2 = aa[1..<4];
println("y2 = #{y2}");
println("yy2 = #{yy2}");

val z2 = 'bc';

assert(y2 == z2);
assert(yy2 == z2);


val y3 = aa[0..10..3];
val yy3 = aa[0..<10..3];
val z3 = 'adgj';
val zz3 = 'adg';

assert(y3 == z3);
assert(yy3 == zz3);

val y4 = a[0..10..3];
val z4 = {1, 4, 7, 10};

assert(y4 == z4);

val xx = 1..20;
val yt = xx[0..10]
val zt = xx[0..<20..2]

assert(yt == [1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
assert(zt == [1, 3, 5, 7, 9, 11, 13, 15, 17, 19])
