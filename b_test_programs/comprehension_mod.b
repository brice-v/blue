# Fixture for test_import_comprehension.b. Comprehensions inside an imported
# module compile their deferred inner program with the module prefix.
val factor = 3;

fun scaled() {
    [i * factor for (i in 1..3)];
}

fun filtered() {
    [i for (i in 1..5) if i % 2 == 0];
}

fun flattened() {
    [x for (row in [[1, 2], [3, 4]]) for (x in row)];
}

fun mapped() {
    {i: i * i for (i in 1..3)};
}

fun setified() {
    {i for (i in 1..3)};
}

fun via_private() {
    [_scale(i) for (i in 1..3)];
}

fun nested() {
    [[j for (j in 1..i)] for (i in 1..3)];
}

fun nested_from_iterable() {
    [i for (i in [j for (j in 1..3)])];
}

fun _scale(i) {
    i * 100;
}

assert(true)