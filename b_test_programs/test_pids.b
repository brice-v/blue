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
    assert(newPid == i+1);
    println("newPid = #{newPid}");
    var newVal = thisPid.recv();
    println("newVal = #{newVal}");
    assert(newVal == i+1);
}

println("thisPid = #{thisPid}");
assert(thisPid == 0);