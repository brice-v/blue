import psutil

# cpu assertions
println('psutil.cpu.percent() = #{psutil.cpu.percent()}');
assert(type(psutil.cpu.percent()) == Type.LIST);
assert(len(psutil.cpu.percent()) != 0);
println('psutil.cpu.info = #{psutil.cpu.info}');
assert(type(psutil.cpu.info) == Type.LIST);
assert(len(psutil.cpu.info) != 0);
println('psutil.cpu.time_info = #{psutil.cpu.time_info}');
assert(type(psutil.cpu.time_info) == Type.LIST);
assert(len(psutil.cpu.time_info) != 0);
println('psutil.cpu.count = #{psutil.cpu.count}');
assert(psutil.cpu.count != 0);

# mem assertions
println('psutil.mem.virtual() = #{psutil.mem.virtual()}');
assert(type(psutil.mem.virtual()) == Type.MAP);
assert(len(psutil.mem.virtual()) != 0);
println('psutil.mem.swap() = #{psutil.mem.swap()}');
assert(type(psutil.mem.swap()) == Type.MAP);
assert(len(psutil.mem.swap()) != 0);

# host assertions
println('psutil.host.info = #{psutil.host.info}');
assert(type(psutil.host.info) == Type.MAP);
assert(len(psutil.host.info) != 0);
# Note: This should work on windows but github runner is throwing an error
if (get_os() != 'windows') {
    println('psutil.host.temps() = #{psutil.host.temps()}');
    assert(type(psutil.host.temps()) == Type.LIST);
    # This also returns nothing when running on gh for linux
    if (get_os() != 'linux') {
        assert(len(psutil.host.temps()) != 0);
    }
}

# net assertions
println('psutil.net.connections() = #{psutil.net.connections()}');
assert(type(psutil.net.connections()) == Type.LIST);
assert(len(psutil.net.connections()) != 0);
println('psutil.net.io_info() = #{psutil.net.io_info()}');
assert(type(psutil.net.io_info()) == Type.LIST);
assert(len(psutil.net.io_info()) != 0);

# disk assertions
println('psutil.disk.partitions = #{psutil.disk.partitions}');
assert(type(psutil.disk.partitions) == Type.LIST);
assert(len(psutil.disk.partitions) != 0);
println('psutil.disk.io_info() = #{psutil.disk.io_info()}')
assert(type(psutil.disk.io_info()) == Type.MAP);
assert(len(psutil.disk.io_info()) != 0);
println('psutil.disk.usage(".") = #{psutil.disk.usage(".")}');
assert(type(psutil.disk.usage(".")) == Type.MAP);
assert(len(psutil.disk.usage(".")) != 0);
