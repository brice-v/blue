# This file is used as a test for importing

fun hello(name="Brice") {
    println("Hello #{name}!")
}

# hello("Something")


fun add(x, y) {
    x + y
}


fun doSomething(some="SomeString") {
    var x = {
        this: "x",
        another: 123,
    };

    return "internal object x is #{x} and some is '#{some}'";
}

fun returnTrue() {
    true
}

true