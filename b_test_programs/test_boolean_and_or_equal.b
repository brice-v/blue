var x = true;

x &&= false;
println(x);
assert(x == false);

x ||= true;
println(x);
assert(x == true);