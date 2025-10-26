package object

import (
	"bytes"
	"os"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/ini"
	"github.com/gookit/config/v2/properties"
	"github.com/gookit/config/v2/toml"
	"github.com/gookit/config/v2/yamlv3"
	"github.com/gookit/ini/v2/dotenv"
)

var ConfigBuiltins = NewBuiltinSliceType{
	{Name: "_load_file", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("load_file", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("load_file", 1, STRING_OBJ, args[0].Type())
			}
			c := config.NewWithOptions("config", config.ParseEnv, config.Readonly)
			c.WithDriver(yamlv3.Driver, ini.Driver, toml.Driver, properties.Driver)
			fpath := args[0].(*Stringo).Value
			err := c.LoadFiles(fpath)
			if err != nil {
				if err.Error() == "not register decoder for the format: env" {
					err = dotenv.LoadFiles(fpath)
					Builtinobjs["ENV"] = &BuiltinObj{
						Obj: PopulateENVObj(),
					}
					if err != nil {
						return newError("`load_file` error: %s", err.Error())
					} else {
						// Need to return a valid JSON value
						return &Stringo{Value: "{}"}
					}
				}
				return newError("`load_file` error: %s", err.Error())
			}
			return &Stringo{Value: c.ToJSON()}
		},
		HelpStr: helpStrArgs{
			explanation: "`load_file` returns the object version of the parsed config file (yaml, ini, toml, properties, json)",
			signature:   "load_file(fpath: str) -> str(json)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "load_file(fpath) => {}",
		}.String(),
	}},
	{Name: "_dump_config", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("dump_config", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("dump_config", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("dump_config", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != STRING_OBJ {
				return newPositionalTypeError("dump_config", 3, STRING_OBJ, args[2].Type())
			}
			c := config.New("config")
			configAsJson := args[0].(*Stringo).Value
			c.LoadStrings(config.JSON, configAsJson)
			fpath := args[1].(*Stringo).Value
			format := args[2].(*Stringo).Value
			out := new(bytes.Buffer)
			switch format {
			case "JSON":
				config.DumpTo(out, config.JSON)
			case "TOML":
				config.DumpTo(out, config.Toml)
			case "YAML":
				config.DumpTo(out, config.Yaml)
			case "INI":
				config.DumpTo(out, config.Ini)
			case "PROPERTIES":
				config.DumpTo(out, config.Prop)
			}
			err := os.WriteFile(fpath, out.Bytes(), 0755)
			if err != nil {
				return newError("`dump_config` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`dump_config` takes the config map and writes it to a file in the given format",
			signature:   "dump_config(c: str(json), fpath: str, format: str('JSON'|'TOML'|'YAML'|'INI'|'PROPERTIES')='JSON') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dump_config(c, 'test.json') => null",
		}.String(),
	}},
}
