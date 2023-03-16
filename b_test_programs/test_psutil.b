import psutil

# cpu assertions
assert(type(psutil.cpu.percent()) == Type.LIST);
assert(len(psutil.cpu.percent()) != 0);
assert(type(psutil.cpu.info) == Type.LIST);
assert(len(psutil.cpu.info) != 0);
assert(type(psutil.cpu.time_info) == Type.LIST);
assert(len(psutil.cpu.time_info) != 0);
assert(psutil.cpu.count != 0);

# mem assertions
assert(type(psutil.mem.virtual()) == Type.MAP);
assert(len(psutil.mem.virtual()) != 0);
assert(type(psutil.mem.swap()) == Type.MAP);
assert(len(psutil.mem.swap()) != 0);
assert(type(psutil.mem.swap_devices()) == Type.LIST);
assert(len(psutil.mem.swap_devices()) != 0);

# host assertions
assert(type(psutil.host.info) == Type.MAP);
assert(len(psutil.host.info) != 0);
assert(type(psutil.host.users()) == Type.LIST);
assert(len(psutil.host.users()) != 0)
assert(type(psutil.host.temps()) == Type.LIST);
assert(len(psutil.host.temps()) != 0);

# net assertions
assert(type(psutil.net.connections()) == Type.LIST);
assert(len(psutil.net.connections()) != 0);
assert(type(psutil.net.io_info()) == Type.LIST);
assert(len(psutil.net.io_info()) != 0);

# disk assertions
assert(type(psutil.disk.partitions) == Type.LIST);
assert(len(psutil.disk.partitions) != 0);
assert(type(psutil.disk.io_info()) == Type.MAP);
assert(len(psutil.disk.io_info()) != 0);
