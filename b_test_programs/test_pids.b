#VM IGNORE
# TODO: Spawn a bunch of things in various ways
# Make sure they all report the right pid
# Make sure they can send/recv from each other

import time

fun thing(parentPid) {
    parentPid.send(self());
    time.sleep(500);
}

var thisPid = self();

for (i in 0..5) {
    println(i);
    var newPid = spawn(thing, [thisPid]);
    assert(newPid.id >= thisPid.id+1);
    println("newPid = #{newPid}");
    var newVal = thisPid.recv();
    println("newVal = #{newVal}");
    assert(newVal.id >= thisPid.id+1);
}

println("thisPid = #{thisPid}");
# Need a UINT here now
assert(thisPid.id >= 0x0);


# What im trying to test here is that sending is non-blocking
# even if we go over the size of 1
#
#
# Buffered channels will not block up to the size of the buffer
# the size is 1 by default so only 1 will go through without blocking
fun listener_thing(parent_pid) {
    var x = self();
    println("x = #{x}");
    for var i = 0; i < 3; i += 1 {
        time.sleep(500);
        var y = x.recv();
        println("y = #{y}");
    }
    parent_pid.send('done');
}

var list_p = spawn(listener_thing, [self()]);
println("list_p = #{list_p}");
list_p.send(1);
list_p.send(2);
list_p.send(3);
println("Done Sending")
val me = self();
var ended = me.recv();
assert(ended == 'done');