## config will allow the user to import a file based configuration
## to be used in programs.
##
## This config can also be exported to a file.
##
## Supported formats are JSON, INI, TOML, YAML, and PROPERTIES


val __load_file = _load_file;
fun load_file(filepath) {
    ##std:this,__load_file
    ## `load_file` takes a filepath and returns a MAP of the configuration
    ##
    ## load_file(filepath: str) -> map[str:str]
    __load_file(filepath).from_json()
}

val __dump_config = _dump_config;
val _acceptable_formats = ["JSON", "YAML", "INI", "TOML", "PROPERTIES"];
fun dump_config(map_to_config, filepath, format="JSON") {
    ##std:this,__dump_config
    ## `dump_config` takes a MAP config and writes it to the given filepath in the set format
    ##
    ## dump_config(map_to_config: map[str:str], filepath: str, format: 'JSON'|'YAML'|'INI'|'TOML'|'PROPERTIES'='JSON') -> null
    val this_config = map_to_config.to_json();
    val upper_format = format.to_upper();
    if (upper_format in _acceptable_formats) {
        __dump_config(this_config, filepath, upper_format)
    } else {
        error("`dump_config` requires format in #{_acceptable_formats}, got=#{format}")
    }
}