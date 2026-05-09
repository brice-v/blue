# Test self(), process operations, and spawn

import time

# Get current process
val me = self()
assert(type(me) == Type.PROCESS)
assert(me.id >= 0)

# Spawn a simple process
fun simpleWorker() {
    return "hello from worker"
}

val pid = spawn(simpleWorker, [])
assert(type(pid) == Type.PROCESS)
assert(pid.id > me.id)

# Wait a bit for the process to complete
time.sleep(100)

# Spawn with arguments
fun adder(a, b) {
    return a + b
}

val addPid = spawn(adder, [10, 20])
time.sleep(100)

# Send and receive
fun receiver(parentPid) {
    val msg = parentPid.recv()
    parentPid.send("received: #{msg}")
}

val recvPid = spawn(receiver, [me])
time.sleep(50)
me.send("hello")
val response = recvPid.recv()
assert(response == "received: hello")

# Multiple sends and receives
var sendCount = 0
var recvCount = 0

fun doubleReceiver(parentPid) {
    for (i in 1..3) {
        val msg = parentPid.recv()
        parentPid.send("echo: #{msg}")
    }
}

val doublePid = spawn(doubleReceiver, [me])
time.sleep(50)

for (i in 1..3) {
    me.send("msg #{i}")
    val resp = doublePid.recv()
    assert(resp == "echo: msg #{i}")
}

# Process with computation
fun compute(n) {
    var sum = 0
    for (i in 1..n) {
        sum += i
    }
    return sum
}

val computePid = spawn(compute, [100])
time.sleep(200)
# The result should be in the channel
# Note: we cant directly get the return value, but we can check the process

# Multiple processes
var results = []
fun countingWorker(id) {
    return id * 10
}

for (i in 1..5) {
    results << spawn(countingWorker, [i])
}

time.sleep(500)
# Results are sent to the parent via channel
for (i in 1..5) {
    val result = me.recv()
    assert(result == i * 10)
}

# Self in nested spawn
fun nestedWorker(parentPid) {
    val childPid = spawn(fun(childParent) {
        childParent.send("from child")
    }, [parentPid])
    time.sleep(100)
    val childMsg = childPid.recv()
    parentPid.send("nested: #{childMsg}")
}

val nestedPid = spawn(nestedWorker, [me])
time.sleep(300)
val nestedResult = me.recv()
assert(nestedResult == "nested: from child")

# Process with error
fun errorWorker() {
    error("worker error")
    return "should not reach"
}

try {
    val errorPid = spawn(errorWorker, [])
    time.sleep(200)
    # The error should be caught somewhere
} catch (e) {
    assert(e.contains("worker error"))
}

# Process with closure
var captured = 0
fun captureWorker(capturedVal) {
    return capturedVal + 1
}

val closurePid = spawn(captureWorker, [41])
time.sleep(100)

# Process with default args
fun defaultWorker(a, b = 10) {
    return a + b
}

val defaultPid = spawn(defaultWorker, [5])
time.sleep(100)

# Verify process id increases
val pid1 = self()
val pid2 = spawn(fun() { return null }, [])
time.sleep(50)
assert(pid2.id > pid1.id)
