for (i in [1,2,3,4,5,6,7]) {
    val y = "Something else"
    var x = "Hello World!"
    println(y,x);
}

for (i in [1,2,3,4,5,6,7]) {
    var y = "Something else"
    if (y == "Something else") {
        y = "Changed";
    }
    var x = "Hello World!"
    println(y,x);
}

assert(true);