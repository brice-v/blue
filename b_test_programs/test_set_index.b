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

if (x1 != 1) {
    return false;
}
println("HERE1");
if (x2 != 10) {
    return false;
}
println("HERE2");
if (z2 != 1) {
    return false;
}
println("HERE3");
if (z3 != 5) {
    return false;
}
println("HERE4");
if ({y for (y in 1..10)}.0 != 1) {
    return false;
}
println("HERE5");

return true;