# Fixture for test_import_selective.b. Only the requested names (plus the private
# helpers they depend on) should be compiled by a selective import.
val _base = 10;

fun _helper(x) {
    x + _base;
}

fun public_double(x) {
    _helper(x) * 2;
}

fun public_unused() {
    "unused";
}

assert(true)
