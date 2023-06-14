val hello = "HELLO";
val x = "     #{hello}     ";

println(len(x));
val centered_hello = hello.center(15);
println(centered_hello);
println(len(centered_hello));
assert(len(x) == len(centered_hello));
assert(x == centered_hello);

val y = "*****#{hello}*****";
val centered_with_stars_hello = hello.center(15, "*");
assert(len(y) == len(centered_with_stars_hello));
assert(y == centered_with_stars_hello);

val z = "12312#{hello}12312";
val centered_with_123_hello = hello.center(15, "123");
assert(len(z) == len(centered_with_123_hello));
assert(z == centered_with_123_hello);