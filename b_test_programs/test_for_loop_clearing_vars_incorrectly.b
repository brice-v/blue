for (i in 0..10) {
    println("i (before) = #{i}");
    for (j in 0..10) {
        println("j (before) = #{j}");
        if (j == 5) {
            break;
        }
        println("j (after) = #{j}")
    }
    try {
        j += 1;
        assert(false, "UNREACHABLE");
    } catch (e) {
        println("GOT HERE e = #{e}");
        assert(e != '`assert` failed: UNREACHABLE');
    }
    println("i (after for) = #{i}");
    assert(i != null, "This confirms i exists");
}

println("SHOULD GET HERE too!")
try {
    i += 1;
    assert(false, "UNREACHABLE");
} catch (e) {
    println("GOT HERE e = #{e}");
    assert(e != '`assert` failed: UNREACHABLE');
}

true;