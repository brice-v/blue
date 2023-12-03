var x = {1,2,3};
var y = {1,2} | {set([1,2,3])}

println("x = #{x}")
println("y = #{y}")
assert(y == {1,2,{1,2,3}})
assert(y != {1,2,{2,2,3}})