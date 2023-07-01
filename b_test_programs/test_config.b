import config

var path_prefix = "b_test_programs/";
try {
    config.load_file("#{path_prefix}test.yml")
} catch (e) {
    # If this failed we may be running from tests where the cwd() is b_test_programs/
    path_prefix = ""
}


val f = config.load_file("#{path_prefix}test.yml");
println("f = #{f}");
println("type(f) = #{type(f)}");
if (f == null) {
    assert(false)
}

val f1 = config.load_file("#{path_prefix}test.ini");
println("f1 = #{f1}");
println("type(f1) = #{type(f1)}");
if (f1 == null) {
    assert(false)
}

val f2 = config.load_file("#{path_prefix}test.toml");
println("f2 = #{f2}");
println("type(f2) = #{type(f2)}");
if (f2 == null) {
    assert(false)
}

val f3 = config.load_file("#{path_prefix}test.json");
println("f3 = #{f3}");
println("type(f3) = #{type(f3)}");
if (f3 == null) {
    assert(false)
}

val f4 = config.load_file("#{path_prefix}test.env");
println("f4 = #{ENV.ENV_TEST}");
println("type(f4) = #{type(f4)}");

if (ENV.ENV_TEST != "blog") {
    assert(false)
}

val f5 = config.load_file("#{path_prefix}test.properties");
println("f5 = #{f5}");
println("type(f5) = #{type(f5)}");
if (f5 == null) {
    assert(false)
}


true;