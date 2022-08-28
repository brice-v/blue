try {
    println("Try");
    # DONT HIT THIS LINE
    assert(false);
} catch (e) {
    println("Catch #{e}");
    #println("Fail with invalid ident #{x}");
} finally {
    println("Finally - Should print after Catch");
}

try {
    println("Try1");
    # DONT HIT THIS LINE
    assert(false);
} catch (e) {
    println("Catch1 #{e}");
}

return true;