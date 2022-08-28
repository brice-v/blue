import abc


abc.hello("Heres a name?")


println(abc.doSomething())

println(abc.add(4,1))

abc.returnTrue()

println("---------------------");

println(abc.useInternalFun());
try {
    println(abc._internalFun());
    println("This should be unreachable");
    return false;
} catch (e) {
    println("Hit exception as expected");
}


println("---------------------");

import foo.bar

return bar.bar();