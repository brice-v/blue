# should be able to use this as a key and it should be a string
var x = {wait: 'hello'}
println(x['wait'])
assert(x['wait'] == 'hello');

var y = {abc: 123};
println(y.type())
assert(y.type() == 'MAP')