## A panic raised by a native builtin must become a blue error, not a process
## crash. ml.matmul with a 1d operand is a borncgo shape error, which borncgo
## raises as a panic; the vm builtin boundary converts it so try/catch runs.

import ml

var caught = false
try {
    val a = ml.tensor([1.0, 2.0, 3.0]);
    val b = ml.tensor([[1.0, 2.0], [3.0, 4.0]]);
    val c = ml.matmul(a, b);
    println(c);
} catch (e) {
    caught = true
} finally {
    assert(caught)
}
assert(caught)
