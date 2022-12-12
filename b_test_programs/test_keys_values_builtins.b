val x = {
    1: "hello",
    'B': 0.203812,
    0x123: [1,2,3],
};

println(x.keys());
assert(x.keys() == [1, 'B', 0x123]);
println(x.values());
assert(x.values() == ['hello', 0.203812, [1,2,3]]);