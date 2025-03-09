import crypto
var original = 1234;

var saved = original.save();
var loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.INT);
assert(type(original) == type(loaded));
assert(original == loaded);

original = 1212424141214212412441214241234n;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.BIGINT);
assert(type(original) == type(loaded));
assert(original == loaded);

original = 1234n;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.BIGINT);
assert(type(original) == type(loaded));
assert(original == loaded);


original = 1234.1234;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.FLOAT);
assert(type(original) == type(loaded));
assert(original == loaded);

original = 1234.1234n;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.BIGFLOAT);
assert(type(original) == type(loaded));
assert(original == loaded);

original = 0x1;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.UINT);
assert(type(original) == type(loaded));
assert(original == loaded);

original = false;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.BOOL);
assert(type(original) == type(loaded));
assert(original == loaded);

original = null;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.NULL);
assert(type(original) == type(loaded));
assert(original == loaded);

original = "HEllo World!";

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.STRING);
assert(type(original) == type(loaded));
assert(original == loaded);

original = [1, 1234n, 1234.124, 1234.1234n, true, false, null, 0x1234, "HELLO WORLD", r/hellow/, "Hello World!".to_bytes()];
val expected_types = [Type.INT, Type.BIGINT, Type.FLOAT, Type.BIGFLOAT, Type.BOOL, Type.BOOL, Type.NULL, Type.UINT, Type.STRING, Type.REGEX, Type.BYTES];

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.LIST);
assert(type(original) == type(loaded));
assert(original == loaded);

assert(len(original) == len(loaded));
assert(len(expected_types) == len(loaded));
for var i = 0; i < expected_types.len(); i += 1 {
    println("type(original[i]) = #{type(original[i])}, expected_types[i] = #{expected_types[i]}")
    assert(type(original[i]) == expected_types[i]);
    assert(type(loaded[i]) == expected_types[i]);
}

original = [x for x in 1..10 if x % 2 == 0];

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.LIST);
assert(type(original) == type(loaded));
assert(original == loaded);

original = r/hellow/;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.REGEX);
assert(type(original) == type(loaded));
assert(original == loaded);

original = "Hello World!".to_bytes();

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.BYTES);
assert(type(original) == type(loaded));
assert(original == loaded);

original = {1, 1234n, 1234.124, 1234.1234n, true, false, null, 0x1234, "HELLO WORLD", r/hellow/, "Hello World!".to_bytes()};

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.SET);
assert(type(original) == type(loaded));
assert(original == loaded);

assert(len(original) == len(loaded));
assert(len(expected_types) == len(loaded));
for var i = 0; i < expected_types.len(); i += 1 {
    println("type(original[i]) = #{type(original[i])}, expected_types[i] = #{expected_types[i]}")
    assert(type(original[i]) == expected_types[i]);
    assert(type(loaded[i]) == expected_types[i]);
}

original = {x for x in 1..10 if x % 2 == 0};

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.SET);
assert(type(original) == type(loaded));
assert(original == loaded);

original = [[1], [1234n, 1234.124], [1234.1234n, true, false], [null], 0x1234, "HELLO WORLD", [r/hellow/, "Hello World!".to_bytes()]];

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.LIST);
assert(type(original) == type(loaded));
assert(original == loaded);

original = {{1}, {1234n, 1234.124}, {1234.1234n, true, false}, {null}, 0x1234, "HELLO WORLD", {r/hellow/, "Hello World!".to_bytes()}};

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.SET);
assert(type(original) == type(loaded));
assert(original == loaded);

original = {{{1}}, 1234, {1234, {1234n, 1234.124}}, {1234, {1}, {1234.1234n, {true}, false}}, {null}, 0x1234, "HELLO WORLD", {r/hellow/, "Hello World!".to_bytes()}};

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.SET);
assert(type(original) == type(loaded));
assert(original == loaded);

original = [[[1]], 1234, [1234, [1234n, 1234.124]], [1234, [1], [1234.1234n, [true], false]], [null], 0x1234, "HELLO WORLD", [r/hellow/, "Hello World!".to_bytes()]];

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.LIST);
assert(type(original) == type(loaded));
assert(original == loaded);

original = {x: 99 for x in 1..10 if x % 2 == 0};

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.MAP);
assert(type(original) == type(loaded));
assert(original == loaded);

original = {x: {99:"Hello"} for x in 1..10 if x % 2 == 0};

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.MAP);
assert(type(original) == type(loaded));
assert(original == loaded);

original = fun() { println("Hello World") };

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.FUN);
assert(type(original) == type(loaded));
assert(original == loaded);

original = fun(a, b, c=1234) { {a: b, c: {1,2}, 'd': "helloworld", 0x123: 1234n} };
val expected_return = {1: 2, 1234: {1,2}, 'd': "helloworld", 0x123: 1234n};

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
assert(type(original) == Type.FUN);
assert(type(original) == type(loaded));
assert(original == loaded);
assert(original(1, 2) == expected_return);
assert(loaded(1, 2) == expected_return);

###
import gg
original = gg.color.light_gray;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
#assert(type(original) == Type.GO_OBJ);
assert(type(original) == type(loaded));
assert(original == loaded);

# With the existing code this wouldnt work due to all members being private
original = FSTDOUT;

saved = original.save();
loaded = saved.load();
println("original = #{original}, loaded = #{loaded}, type(original) = #{type(original)}, type(loaded) = #{type(loaded)} saved = #{crypto.encode(saved)}");
#assert(type(original) == Type.GO_OBJ);
assert(type(original) == type(loaded));
assert(original == loaded);
###

import math

original = math;

try {
    saved = original.save();
    assert(false);
} catch (e) {
    println(e);
    assert(e == "EvaluatorError: `save` error: MODULE_OBJ is not supported for encoding")
}

assert(true);