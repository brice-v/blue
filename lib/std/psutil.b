## `psutil` contains process and system utilities

val __cpu_info = _cpu_info;
val __cpu_time_info = _cpu_time_info;
val __mem_virt_info = _mem_virt_info;
val __mem_swap_info = _mem_swap_info;
val __mem_swap_devices = _mem_swap_devices;

fun _cpu_info_json_to_map() {
    var __result = [];
    for (__e in __cpu_info()) {
        __result << __e.from_json();
    }
    return __result;
}

fun _cpu_time_info_json_to_map() {
    var __result = [];
    for (__e in __cpu_time_info()) {
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
    for (__e in __mem_swap_devices()) {
        __result << __e.from_json();
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
