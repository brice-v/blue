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
val response = me.recv()
assert(response == "received: hello")

# Multiple sends and receives
var sendCount = 0
var recvCount = 0

fun doubleReceiver(parentPid) {
    val meHere = self();
    for (i in 1..3) {
        val msg = meHere.recv()
        parentPid.send("echo: #{msg}")
    }
}

val doublePid = spawn(doubleReceiver, [me])
time.sleep(50)

for (i in 1..3) {
    doublePid.send("msg #{i}")
    val resp = me.recv()
    assert(resp == "echo: msg #{i}")
}

# Process with computation
fun compute(n, parentPid) {
    var sum = 0
    for (i in 1..n) {
        sum += i
    }
    parentPid.send(sum);
}

val computePid = spawn(compute, [100, me])
time.sleep(200)
val computeResult = me.recv();
assert(computeResult == 5050)

# Multiple processes
var results = []
fun countingWorker(id, parentPid) {
    parentPid.send(id * 10)
}

for (i in 1..5) {
    results << spawn(countingWorker, [i, me])
}

time.sleep(500)
# Results are sent to the parent via channel
for (i in 1..5) {
    val result = me.recv()
    assert(result == i * 10)
}

# Self in nested spawn
fun nestedWorker(parentPid) {
    val childPid = spawn(fun() {
        self().send("from child")
    }, [])
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
    self().send(capturedVal + 1)
}

val closurePid = spawn(captureWorker, [41])
time.sleep(100)
assert(closurePid.recv() == 42);

# Process with default args
fun defaultWorker(a, b = 10) {
    self().send(a + b)
}

val defaultPid = spawn(defaultWorker, [5])
time.sleep(100)
assert(defaultPid.recv() == 15);

# Verify process id increases
val pid1 = self()
val pid2 = spawn(fun() { return null }, [])
time.sleep(50)
assert(pid2.id > pid1.id)
