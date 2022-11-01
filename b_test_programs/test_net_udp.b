import net

fun something(parent_pid) {
    println("Spawning Listener...");
    val c = net.listen(transport="udp");
    println("c = #{c}");

    var x = c.read();
    println("x = #{x}");
    parent_pid.send(true);
}

val parent_pid = self();

val child_pid = spawn(something, [parent_pid]);

println("parent_pid = #{parent_pid}, child_pid = #{child_pid}");

import time
for (true) {
    try {
        val connection = net.connect(transport="udp");
        println("connection = #{connection}");
        try {
            connection.write("SOMETHING!!!!");
        } catch (e) {
            println("error: #{e}");
            break;
        }
        
    } catch (e) {
        continue;
    }

    val x = parent_pid.recv()
    if (x) {
        println("DONE");
        return null;
    }
    time.sleep(100);
}

return true;
