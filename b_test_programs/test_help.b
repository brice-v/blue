## This module tests the `help` function


fun main() {
    ## `main` is the entry point of this application
    ## more here
    ""
}



assert(main() == "");

println(help(main));
val expected_help = """`main` is the entry point of this application
more here

type(main) = 'FUNCTION'
inspect(main) = 'fun() {
""
}'""".replace("\r", "");
assert(help(main) == expected_help);

## `on_this` is a random variable that we 
## want to see if this gets picked up
var on_this = 1;
assert(true);

#println(help(on_this));


# TODO: So what we MAYBE could do is 'quote' it like we do for spawn?
# - maybe the default for help() could be returing the module's help?
import config
val config_help = help(config);
println(config_help);
val expected_config_help = """MODULE `config`: config will allow the user to import a file based configuration
to be used in programs.

This config can also be exported to a file.

Supported formats are JSON, INI, TOML, YAML, and PROPERTIES

type(config) = 'MODULE_OBJ'

PUBLIC FUNCTIONS:
load_file | `load_file` takes a filepath and returns a MAP of the configuration
            
            load_file(filepath: str) -> map[str:str]
dump_config | `dump_config` takes a MAP config and writes it to the given filepath in the set format
              
              dump_config(map_to_config: map[str:str], filepath: str, format: 'JSON'|'YAML'|'INI'|'TOML'|'PROPERTIES'='JSON) -> null""".replace("\r", "");
assert(config_help == expected_config_help)