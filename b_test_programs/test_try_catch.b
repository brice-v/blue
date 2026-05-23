try {
    println("Try");
    assert(false);
} catch (e) {
    println("Catch #{e}");
    assert(e == "`assert` failed");
} finally {
    println("Finally - Should print after Catch");
}

try {
    println("Try1");
    assert(false);
} catch (e) {
    println("Catch1 #{e}");
    assert(e == "`assert` failed");
}


assert(true);