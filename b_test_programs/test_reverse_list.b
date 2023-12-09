val x = ['a','n',1,0x01,{'abc':123},{4,5,6},[90]];
println("x = #{x}");
val x_rev = x.reverse();
println("x_rev = #{x_rev}");
val x_rev_expected = [[90],{4,5,6},{'abc':123},0x01,1,'n','a'];
println("x_rev_expected = #{x_rev_expected}");
assert(x == ['a','n',1,0x01,{'abc':123},{4,5,6},[90]]);
assert(x_rev == x_rev_expected);