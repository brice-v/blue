var x = {1,3,4};

x << 1;
println("x = #{x}")
assert(x == {1,3,4});
x << 'a';
println("x = #{x}")
assert(x == {1,3,4,'a'});

1 >> x;
println("x = #{x}")
assert(x == {1,3,4,'a'});
'a' >> x;
println("x = #{x}")
assert(x == {1,3,4,'a'});
6 >> x;
println("x = #{x}")
assert(x == {1,3,4,'a',6});

x += 1;
println("x = #{x}")
assert(x == {1,3,4,'a',6});
x += 'a';
println("x = #{x}")
assert(x == {1,3,4,'a',6});

x += 5;
println("x = #{x}")
assert(x == {1,3,4,'a',6,5});