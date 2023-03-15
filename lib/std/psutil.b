## `psutil` contains process and system utilities

val __cpu_info = _cpu_info;
val __cpu_time_info = _cpu_time_info;
val __mem_virt_info = _mem_virt_info;
val __mem_swap_info = _mem_swap_info;
val __mem_swap_devices = _mem_swap_devices;
val __host_info = _host_info;
val __host_users_info = _host_users_info;
val __host_temps_info = _host_temps_info;
val __net_connections = _net_connections;
val __net_io_info = _net_io_info;
val __disk_partitions = _disk_partitions;
val __disk_io_info = _disk_io_info;

fun _cpu_info_json_to_map() {
    var __result = [];
    val ___data = __cpu_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun _cpu_time_info_json_to_map() {
    var __result = [];
    val ___data = __cpu_time_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun _mem_info_to_map() {
    __mem_virt_info().from_json()
}

fun _mem_swap_to_map() {
    __mem_swap_info().from_json()
}

fun _mem_swap_devices_to_list_of_maps() {
    var __result = [];
    val ___data = __mem_swap_devices();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun _host_info_to_map() {
    __host_info().from_json()
}

fun _host_users_info_to_map() {
    var __result = [];
    val ___data = __host_users_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun _host_temps_info_to_map() {
    var __result = [];
    val ___data = __host_temps_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun _net_connections_to_map(option="all") {
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

fun _net_io_info_to_map() {
    var __result = [];
    val ___data = __net_io_info();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun _disk_partions_to_map() {
    var __result = [];
    val ___data = __disk_partitions();
    for (__e in ___data) {
        __result << __e.from_json();
    }
    return __result;
}

fun _disk_io_info_to_map() {
    var __result = {};
    val ___data = __disk_io_info();
    for ([__k,__e] in ___data) {
        __result[__k] = __e.from_json();
    }
    return __result;
}

val cpu = {
    percent: _cpu_usage_percent,
    info: _cpu_info_json_to_map(),
    time_info: _cpu_time_info_json_to_map(),
    count: _cpu_count(),
};

val mem = {
    virtual: _mem_info_to_map,
    swap: _mem_swap_to_map,
    swap_devices: _mem_swap_devices_to_list_of_maps,
};

val host = {
    info: _host_info_to_map(),
    users: _host_users_info_to_map,
    temps: _host_temps_info_to_map,
};

val net = {
    connections: _net_connections_to_map,
    io_info: _net_io_info_to_map,
};

val disk = {
    partitions: _disk_partions_to_map(),
    io_info: _disk_io_info_to_map,
};