import config

var path_prefix = "";

try {
    config.load_file("#{path_prefix}test.yml")
} catch (e) {
    # If this failed we may be running from tests where the cwd() is b_test_programs/
    path_prefix = "b_test_programs/"
}


val f = config.load_file("#{path_prefix}test.yml");
println("f = #{f}");
println("type(f) = #{type(f)}");
assert(f != null);

val f1 = config.load_file("#{path_prefix}test.ini");
println("f1 = #{f1}");
println("type(f1) = #{type(f1)}");
assert(f1 != null);

###
val f2 = config.load_file("#{path_prefix}test.toml");
println("f2 = #{f2}");
println("type(f2) = #{type(f2)}");
assert(f2 != null);

val f3 = config.load_file("#{path_prefix}test.json");
println("f3 = #{f3}");
println("type(f3) = #{type(f3)}");
assert(f3 != null);

val f4 = config.load_file("#{path_prefix}test.env");
println("f4 = #{ENV.ENV_TEST}");
println("type(f4) = #{type(f4)}");

assert(ENV.ENV_TEST == "blog");

val f5 = config.load_file("#{path_prefix}test.properties");
println("f5 = #{f5}");
println("type(f5) = #{type(f5)}");
assert(f5 != null);
