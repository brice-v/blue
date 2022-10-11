val load_file_ = _load_file;
fun load_file(filepath) {
    load_file_(filepath).json_to_map()
}