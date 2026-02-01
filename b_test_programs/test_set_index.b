val x = {y for (y in 1..10)};
val z = {1, 2, 3, 4, 5}

val x1 = x[0];
# This is not possible with current parser (not planning on supporting it)
#val x2 = x.x.len();
val x2 = x[len(x)-1];
val z2 = z.0;
val z3 = z[z.len()-1]
#println("len(x)=#{len(x)}, z.len()=#{z.len()}");
#println("x=#{x}, z=#{z}, x1=#{x1}, x2=#{x2}, z2=#{z2}, z3=#{z3}");

assert(x1 == 1);
println("HERE1");
assert(x2 == 10);
println("HERE2");
assert(z2 == 1)
println("HERE3");
assert(z3 == 5)
println("HERE4");
assert({y for (y in 1..10)}.0 == 1)
println("HERE5");

assert(true);