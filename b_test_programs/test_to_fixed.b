#VM IGNORE
# JS-style to_fixed(digits) returns a fixed-point STRING.
val a = 1.2345;
val b = 1.2;
val c = -1.5;
val d = 1.239;
val e = 1.6;
val i = 2;

assert(a.to_fixed(2) == "1.23");
assert(b.to_fixed(2) == "1.20");
assert(c.to_fixed(2) == "-1.50");
assert(d.to_fixed(2) == "1.24");
assert(e.to_fixed(0) == "2");
assert(i.to_fixed(2) == "2.00");
assert(b.to_fixed(3) == "1.200");
println("to_fixed ok");
