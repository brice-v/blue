val x = [1,2,3,4,''];

fun do_something(l) {
    var rv = null;
    for (a in l) {
        if (a == '') {
            continue;
        }
        rv = a;
    }
    return rv;
}

println("do_something(x) = #{do_something(x)}");
assert(do_something(x) == 4);