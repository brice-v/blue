import ml

var a = ml.tensor([[1, 2],[3, 4]])
var b = ml.tensor([[2, 0],[1, 2]])
var expected = ml.tensor([[4,4],[10,8]])
a @= b
println(a)
assert(ml.equal(a, expected))