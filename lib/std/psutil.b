## `psutil` contains process and system utilities
##
## The following constants are defined with
## relative functions for it
## If the data is defined as a map it is likely to
## be static - but this could change
##
## cpu  | percent, info, time_info, count
##      | cpu.percent() -> list[float]
##      | cpu.info -> list[map[str:any]]
##      | cpu.time_info -> list[map[str:any]]
##      | cpu.count() -> int
## ----------------------------------------------
## mem  | virtual, swap, swap_devices
##      | mem.virtual() -> map[str:int]
##      | mem.swap() -> map[str:int|float]
## ----------------------------------------------
## host | info, users, temps
##      | host.info -> map[str:any]
##      | host.temps() -> list[map[str:any]]
## ----------------------------------------------
## net  | connections, io_info
##      | net.connections(option: str='all') ->
##      | list[map[str:any]]
##      | net.io_info() -> list[map[str:any]]
## ----------------------------------------------
## disk | partitions, io_info
##      | disk.partitions -> list[map[str:any]]
##      | disk.io_info() -> map[str:any]
##      | disk.usage(path: str) -> map[str:any]

val __cpu_info = _cpu_info;
val __cpu_time_info = _cpu_time_info;
val __mem_virt_info = _mem_virt_info;
val __mem_swap_info = _mem_swap_info;
val __host_info = _host_info;
val __host_temps_info = _host_temps_info;
val __net_connections = _net_connections;
val __net_io_info = _net_io_info;
val __disk_partitions = _disk_partitions;
val __disk_io_info = _disk_io_info;
val __disk_usage = _disk_usage;

fun psutil_cpu_info_json_to_map() {
    ##std:this,__cpu_info
    ## `cpu.info`: `psutil_cpu_info_json_to_map` returns the mapped version of cpu_info json
    ## 
    ## cpu.info -> list[map[str:any]]
    var __result = [];
    val ___data = __cpu_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun psutil_cpu_time_info_json_to_map() {
    ##std:this,__cpu_time_info
    ## `cpu.time_info`: `psutil_cpu_time_info_json_to_map` returns the mapped version of cpu_time_info json
    ## 
    ## cpu.time_info -> list[map[str:any]]
    var __result = [];
    val ___data = __cpu_time_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun psutil_mem_info_to_map() {
    ##std:this,__mem_virt_info
    ## `mem.virtual()`: `psutil_mem_info_to_map` returns the mapped version of mem_virt_info json
    ## 
    ## mem.virtual() -> map[str:int]
    __mem_virt_info().from_json()
}

fun psutil_mem_swap_to_map() {
    ##std:this,__mem_swap_info
    ## `mem.swap()`: `psutil_mem_swap_to_map` returns the mapped version of mem_swap_info json
    ## 
    ## mem.swap() -> map[str:int|float]
    __mem_swap_info().from_json()
}

fun psutil_host_info_to_map() {
    ##std:this,__host_info
    ## `host.info`: `psutil_host_info_to_map` returns the mapped version of host_info json
    ## 
    ## host.info -> map[str:any]
    __host_info().from_json()
}

fun psutil_host_temps_info_to_map() {
    ##std:this,__host_temps_info
    ## `host.temps()`: `psutil_host_temps_info_to_map` returns the mapped version of host_temps_info json
    ## 
    ## host.temps() -> list[map[str:any]]
    var __result = [];
    val ___data = __host_temps_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun psutil_net_connections_to_map(option="all") {
    ##std:this,__net_connections
    ## `net.connections()`: `psutil_net_connections_to_map` returns the mapped version of net_connections json
    ## 
    ## net.connections(option: str='all') -> list[map[str:any]]
    val __valid_options = ["all", "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6", "unix", "inet", "inet4", "inet6"];
    if (option notin __valid_options) {
        return error("expected a valid option: #{__valid_options}");
    }
    val ___conns = __net_connections(option);
    var __result = [];
    for (__e in ___conns) {
        __result << __e.from_json();
    }
    return __result;
}

fun psutil_net_io_info_to_map() {
    ##std:this,__net_io_info
    ## `net.io_info()`: `psutil_net_io_info_to_map` returns the mapped version of net_io_info json
    ## 
    ## net.io_info() -> list[map[str:any]]
    var __result = [];
    val ___data = __net_io_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun psutil_disk_partions_to_map() {
    ##std:this,__disk_partitions
    ## `disk.partitions`: `psutil_disk_partions_to_map` returns the mapped version of disk_partitions json
    ## 
    ## disk.partitions -> list[map[str:any]]
    var __result = [];
    val ___data = __disk_partitions();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun psutil_disk_io_info_to_map() {
    ##std:this,__disk_io_info
    ## `disk.io_info()`: `psutil_disk_io_info_to_map` returns the mapped version of disk_io_info json
    ## 
    ## disk.io_info() -> map[str:any]
    var __result = {};
    val ___data = __disk_io_info();
    for ([__k,__e] in ___data) {
        __result[__k] = __e.from_json();
    }
    return __result;
}

fun psutil_disk_usage_to_map(path) {
    ##std:this,__disk_io_info
    ## `disk.usage(path)`: `psutil_disk_usage_to_map` returns the mapped version of disk_usage json
    ## 
    ## disk.usage(path: str) -> map[str:any]
    __disk_usage(path).from_json()
}

val cpu = {
    percent: _cpu_usage_percent,
    info: psutil_cpu_info_json_to_map(),
    time_info: psutil_cpu_time_info_json_to_map(),
    count: _cpu_count(),
};

val mem = {
    virtual: psutil_mem_info_to_map,
    swap: psutil_mem_swap_to_map,
};

val host = {
    info: psutil_host_info_to_map(),
    temps: psutil_host_temps_info_to_map,
};

val net = {
    connections: psutil_net_connections_to_map,
    io_info: psutil_net_io_info_to_map,
};

val disk = {
    partitions: psutil_disk_partions_to_map(),
    io_info: psutil_disk_io_info_to_map,
    usage: psutil_disk_usage_to_map,
};