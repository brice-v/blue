## `psutil` contains process and system utilities

val __cpu_info = _cpu_info;
val __cpu_time_info = _cpu_time_info;

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

val cpu = {
    percent: _cpu_usage_percent,
    info: _cpu_info_json_to_map(),
    time_info: _cpu_time_info_json_to_map(),
    count: _cpu_count(),
};