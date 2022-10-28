val __load_file = _load_file;
fun load_file(filepath) {
    __load_file(filepath).json_to_map()
}

val __dump_config = _dump_config;
val _acceptable_formats = ["JSON", "YAML", "INI", "TOML", "PROPERTIES"];
fun dump_config(map_to_config, filepath, format="JSON") {
    val this_config = map_to_config.to_json();
    val upper_format = format.to_upper();
    if (upper_format in _acceptable_formats) {
        __dump_config(filepath, upper_format)
    } else {
        error("`dump_config` requires format in #{_acceptable_formats}, got=#{format}")
    }
}