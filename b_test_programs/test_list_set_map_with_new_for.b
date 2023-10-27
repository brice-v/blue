val abc = [x for var x = 0; x < 10; x += 1];
println("abc = #{abc}");
val abc1 = [x for (var x = 0; x < 10; x += 1)];
println("abc1 = #{abc1}");
val abc2 = {x for (var x = 0; x < 10; x += 1)};
println("abc2 = #{abc2}");
val abc3 = {x for var x = 0; x < 10; x += 1};
println("abc3 = #{abc3}");
val abc4 = {x: 'a' for var x = 0; x < 10; x += 1};
println("abc4 = #{abc4}");
val abc5 = {x: 'a' for (var x = 0; x < 10; x += 1)};
println("abc5 = #{abc5}");

assert(abc == [0,1,2,3,4,5,6,7,8,9])
assert(abc1 == [0,1,2,3,4,5,6,7,8,9])
assert(abc2 == {0,1,2,3,4,5,6,7,8,9})
assert(abc3 == {0,1,2,3,4,5,6,7,8,9})
assert(abc4 == {0:'a',1:'a',2:'a',3:'a',4:'a',5:'a',6:'a',7:'a',8:'a',9:'a'})
assert(abc5 == {0:'a',1:'a',2:'a',3:'a',4:'a',5:'a',6:'a',7:'a',8:'a',9:'a'})