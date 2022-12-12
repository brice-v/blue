
# X is list, Y is expr.
# X << Y append (or popfront if assigned to var)
# Y >> X push (or popback if assigned to var)

var new_list = fun() {
    [1, 2, 3]
}
var x1 = new_list();
var x2 = new_list();
var x3 = new_list();
var x4 = new_list();
var y1 = 1;
var y2 = "Y";
var y3 = {'hello': 'world'};

var a = << x1;
println("a = #{a}, x1 = #{x1}");
assert(a == 1);
assert(x1 == [2,3]);
var b = x2 >>;
println("b = #{b}, x2 = #{x2}");
assert(b == 3);
assert(x2 == [1,2]);
var c = x3 << y1;
println("c = #{c}, x3 = #{x3}, y1 = #{y1}")
assert(c == null);
assert(x3 == [1,2,3,1]);
var d = y2 >> x4;
println("d = #{d}, x4 = #{x4}, y2 = #{y2}");
assert(d == null);
assert(x4 == ['Y',1,2,3]);

println("a = #{a}, b = #{b}, c = #{c}, d = #{d}");
println("x1 = #{x1}, x2 = #{x2}, x3 = #{x3}, x4 = #{x4}");

true