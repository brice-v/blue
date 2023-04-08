val x = null or "Hello World";
println(x);
assert(x == "Hello World");

val y = null;

val z = y or "Another";
println(z);
assert(z == "Another");

val a = "Something";
val b = null;
val c = b or a;

assert(c == a);