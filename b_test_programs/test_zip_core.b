var l1 = [1,2,3, 'a'];
var l2 = [{}, null, 'blah'];
var l3 = [6,7,8,8,9,'c'];

var expected = [[1,{},6],[2,null,7],[3,'blah',8]];
var actual = zip([l1,l2,l3]);
println(actual);
assert(expected == actual);

var left = [1, 1, 3, 1, 1];
var right = [1, 1, 5, 1, 1];
actual = zip([left, right]);
expected = [[1,1],[1,1],[3,5],[1,1],[1,1]];
println(actual);
assert(expected == actual);