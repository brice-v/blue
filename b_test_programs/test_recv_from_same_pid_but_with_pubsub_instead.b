# So we are not going to support sending on a channel to more than 1 at the same time
# and instead pubsub will be a valid option

import time
fun do_something(topic) {
    val my_pid = self();
    val sub = pubsub.subscribe(topic)
    println("Spawned Process on PID = #{my_pid}, subscribed on topic #{topic}");
    for (true) {
        val x = sub.recv();
        println("x = #{x}, self = #{my_pid}");
        return null;
    }
}

val me = 'me';

val pid1 = spawn(do_something, [me]);
val pid2 = spawn(do_something, [me])

for (true) {
    if (pubsub.get_subscriber_count() != 2 ) {
        time.sleep(100);
        continue;
    }
    break;
}
println("GOT HERE")
# Support recv from multiple processes
pubsub.publish(me, "Hello");

wait(pid1, pid2);

assert(true)
