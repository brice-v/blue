var expected = {'a': {'b': {'a': {}}}};
println("expected = #{expected}")
#return true;

var x = {'a': {}};

x.a['b'] = new(x);
# this should be the equivalent of saying {'a': {'b': {'a': {}}}}

println(x);
assert(true);