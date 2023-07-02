val hello = "HELLO";
val x = "     #{hello}     ";

println(len(x));
val centered_hello = hello.center(15);
println(centered_hello);
println(len(centered_hello));
assert(len(x) == len(centered_hello));
assert(x == centered_hello);

val y = "*****#{hello}*****";
val centered_with_stars_hello = hello.center(15, "*");
assert(len(y) == len(centered_with_stars_hello));
assert(y == centered_with_stars_hello);

val z = "12312#{hello}12312";
val centered_with_123_hello = hello.center(15, "123");
assert(len(z) == len(centered_with_123_hello));
assert(z == centered_with_123_hello);

val right_justified = "     HELLO";
assert(hello.rjust(10) == right_justified);
val left_justified = "HELLO     ";
assert(hello.ljust(10) == left_justified);

val right_justified1 = "12312HELLO";
assert(hello.rjust(10, pad="123") == right_justified1);
val left_justified1 = "HELLO12312";
assert(hello.ljust(10, "123") == left_justified1);

val padded_lr_hello = " #{hello} ";
assert(padded_lr_hello.rstrip() == " #{hello}");
assert(padded_lr_hello.lstrip() == "#{hello} ");

assert(hello.reverse() == "OLLEH");
assert(hello.to_title() == "Hello");
assert("hello-world".to_camel() == "HelloWorld");
assert("Hello-World".to_snake() == "hello_world");
assert("Hello World".to_kebab() == "hello-world");


val example_string_to_match = "Some odd string with another woRd";
val regex_to_use = ".*\\swoRd$";
assert(example_string_to_match.matches(regex_to_use));
val example_string_to_not_match = "Some odd string with anotherwoRd";
assert(!example_string_to_not_match.matches(regex_to_use));