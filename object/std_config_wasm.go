//go:build wasm

package object

// ConfigBuiltins for wasm builds. The gookit/config dependency chain does
// not compile for js/wasm, so the config builtins report a clear runtime
// error instead.
var ConfigBuiltins = []*Builtin{
	{
		Name: "_load_file",
		Fun: func(args ...Object) Object {
			return newError("`load_file` error: the config module is not available in the wasm build of blue")
		},
		HelpStr: helpStrArgs{
			explanation: "`load_file` is not available in the wasm build of blue",
			signature:   "load_file(fpath: str) -> Error",
			errors:      "CustomError",
			example:     "load_file(fpath) => 'load_file error: the config module is not available in the wasm build of blue'",
		}.String(),
	},
	{
		Name: "_dump_config",
		Fun: func(args ...Object) Object {
			return newError("`dump_config` error: the config module is not available in the wasm build of blue")
		},
		HelpStr: helpStrArgs{
			explanation: "`dump_config` is not available in the wasm build of blue",
			signature:   "dump_config(config_as_json: str, fpath: str, format: str) -> Error",
			errors:      "CustomError",
			example:     "dump_config('{}', './out.json', 'JSON') => 'dump_config error: the config module is not available in the wasm build of blue'",
		}.String(),
	},
}
