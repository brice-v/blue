try {
    {'a': 1}.x();
} catch (e) {
    assert(e == "not a function NULL, index `x` is not in environment" || e == "calling non-closure and non-builtin. got=NULL")
}

{'a': 1, 'x': fun() {println("Hello World!")}}.x();

assert(true);