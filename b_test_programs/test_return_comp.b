fun explicit_return_list() {
    return [x for (x in 0..10)];
}

fun implicit_return_list() {
    [x for (x in 0..10)]
}

fun explicit_return_set() {
    return {x for (x in 0..10)};
}

fun implicit_return_set() {
    {x for (x in 0..10)}
}

fun explicit_return_map() {
    return {x:x for (x in 0..10)};
}

fun implicit_return_map() {
    {x:x for (x in 0..10)}
}

println("type(explicit_return_list()) = #{type(explicit_return_list())}");
var explicit_list = explicit_return_list()[0];
println("explicit_list = #{explicit_list}");
println("type(implicit_return_list()) = #{type(implicit_return_list())}");
var implicit_list = implicit_return_list()[0];
println("implicit_list = #{implicit_list}");
println("type(explicit_return_set()) = #{type(explicit_return_set())}");
var explicit_set = explicit_return_set()[0];
println("explicit_set = #{explicit_set}");
println("type(implicit_return_set()) = #{type(implicit_return_set())}");
var implicit_set = implicit_return_set()[0];
println("implicit_set = #{implicit_set}");
println("type(explicit_return_map()) = #{type(explicit_return_map())}");
var explicit_map = explicit_return_map()[0];
println("explicit_map = #{explicit_map}");
println("type(implicit_return_map()) = #{type(implicit_return_map())}");
var implicit_map = implicit_return_map()[0];
println("implicit_map = #{implicit_map}");

true