fun add() {
    for (true) {
        var x = recv();
        if (x == "Hello") {
            println("Hello World!");
            return null;
        } else {
            println("Nothing recv #{x}");
        }
    }
}

var pid = spawn(add);

println("self() = #{self()}, pid = #{pid}");

#var pid = self();


pid.send("Something else");
pid.send("Hello");

import time

time.sleep(1000)