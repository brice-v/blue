from abc import {hello, doSomething, add, returnTrue, useInternalFun}


hello("Heres a name?")


println(doSomething())

println(add(4,1))

returnTrue()

println("---------------------");

println(useInternalFun());
### This is a compiler error for VM
try {
    println(_internalFun());
    println("This should be unreachable");
    return false;
} catch (e) {
    println("Hit exception as expected");
}
###


println("-----------------521152----");

from foo.bar import *

val xyzabc = bar() and false;
println("----------------------------bar = #{xyzabc}");
assert(bar());

import foo.bar as b

return b.bar();