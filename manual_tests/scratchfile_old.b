## `scratchfile` is a file used for testing


###
import color
import math
import time
#import test_parser_error

var s = color.style(color.italic, color.magenta)
println(s, "italic")

var xyz = {};
xyz.print = 'something';
print(xyz);


fun hello(x, y = 3+2, z = 4, a) {
    ## `hello` is a function I used to test default args
    val duration = math.rand() * 10 * 100;
    println("sleeping for #{duration}ms");
    time.sleep(int(duration))
    return x + y + z + a;
}

assert(hello(3,5) == 17)

var pids = [];
for (x in 1..10) {
    #println("in loop, x = #{x}")
    pids << spawn(hello, [1,2,3,4]);
}
println("pids = #{pids}")

wait(pids)

# This block is all good too
fun different() {
    val duration = int(math.rand() * 10 * 100);
    println("sleeping for #{duration}ms");
    time.sleep(duration)
}

pids = [];
for (x in 1..10) {
    pids << spawn(different);
}
println("second pids = #{pids}");
wait(pids);


var x = {};
x.name = {};
x.name.hello = 'thing';
println(x);



import psutil

for (_ in 1..10) {
    val usage = psutil.cpu.percent();
    println("usage = #{usage}");
    time.sleep(100);
}

for ([k,v] in psutil.cpu.info[0]) {
    print(s, k);
    println(": #{v}");
}

println("psutil.cpu.time_info = #{psutil.cpu.time_info}")
println("psutil.cpu.count = #{psutil.cpu.count}")

println("HERE")

###


#wasm.init('add.wasm');
println(get_os())
println(get_arch())

KV.put('hello', 'os', get_os());
val x = KV.get('hello', 's');

if (x) {
    println("x = #{x}")
}