fun hello(arg1=10, arg2="", arg3=true) {
    val rv = "arg1 = #{arg1}, arg2 = #{arg2}, arg3 = #{arg3}";
    println(rv);
    return rv;
}


var result = hello(1, "Hello", false);
var expected = "arg1 = 1, arg2 = Hello, arg3 = false";
assert(result == expected);

result = hello(arg2="Two");
expected = "arg1 = 10, arg2 = Two, arg3 = true"
assert(result == expected);

result = hello(arg3=false);
expected = "arg1 = 10, arg2 = , arg3 = false";
assert(result == expected);

result = hello(arg1=100, "something");
expected = "arg1 = 100, arg2 = something, arg3 = true";
assert(result == expected);

result = hello(1010, arg2="another");
expected = "arg1 = 1010, arg2 = another, arg3 = true";
assert(result == expected);