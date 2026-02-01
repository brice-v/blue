val hw = "Hello World";
val h = "Hello";
val w = "World";

assert("Hello World".split(" ") == ["Hello", "World"]);
assert("Hello World".split(" ") == [h, w]);
assert(hw.split(" ") == ["Hello", "World"]);
assert(hw.split(" ") == [h, w]);
assert(hw.split() == [h, w]);
assert(split("Hello World") == ["Hello", "World"]);
assert(split("Hello World", " ") == ["Hello", "World"]);
assert(split(hw, " ") == ["Hello", "World"]);
assert(split(hw) == [h, "World"]);
assert(split("Hello World") == [h, w]);
assert(split("Hello World", " ") == ["Hello", w]);
assert(split(hw, " ") == [h, "World"]);
assert(split(hw) == [h, w]);