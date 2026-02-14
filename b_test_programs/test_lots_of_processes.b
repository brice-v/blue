#VM IGNORE
# TODO: vm will support processes

import time
fun read_and_exit() {
	val current_pid = self();
	#println("self = #{current_pid}");
	for (true) {
		time.sleep(100);
		val value = current_pid.recv();
		if (value != null) {
			println("current_pid = #{current_pid}, value = #{value}");		
			return null;
		}
	}
}

val self_pid = self();
println("self_pid = #{self_pid}");
var pids = [];
# test manually with more to make sure but this adds a fair amount of time to
# the test cases - like 10s (or 2x previous time really)
for (i in 0..500) {
	pids << spawn(read_and_exit, []);
}
println("pids = #{pids}")

for (pid in pids) {
	println("pid = #{pid}");
	pid.send("Hello to pid #{pid}");
}

assert(true);