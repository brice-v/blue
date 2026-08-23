//go:build wasm

package object

// DbBuiltins for wasm builds. The pure-go sqlite driver does not compile
// for js/wasm, and browsers sandbox away local databases anyway, so every
// db builtin reports a clear runtime error instead.
var DbBuiltins = []*Builtin{
	{
		Name: "_db_open",
		Fun: func(args ...Object) Object {
			return newError("`db_open` error: the db module is not available in the wasm build of blue")
		},
		HelpStr: helpStrArgs{
			explanation: "`db_open` is not available in the wasm build of blue",
			signature:   "db_open(db_name: str=':memory:') -> Error",
			errors:      "CustomError",
			example:     "db_open() => 'db_open error: the db module is not available in the wasm build of blue'",
		}.String(),
	},
	{
		Name: "_db_ping",
		Fun: func(args ...Object) Object {
			return newError("`db_ping` error: the db module is not available in the wasm build of blue")
		},
		HelpStr: helpStrArgs{
			explanation: "`db_ping` is not available in the wasm build of blue",
			signature:   "db_ping(db: GoObj[*sql.DB]) -> Error",
			errors:      "CustomError",
			example:     "db_ping(db) => 'db_ping error: the db module is not available in the wasm build of blue'",
		}.String(),
	},
	{
		Name: "_db_close",
		Fun: func(args ...Object) Object {
			return newError("`db_close` error: the db module is not available in the wasm build of blue")
		},
		HelpStr: helpStrArgs{
			explanation: "`db_close` is not available in the wasm build of blue",
			signature:   "db_close(db: GoObj[*sql.DB]) -> Error",
			errors:      "CustomError",
			example:     "db_close(db) => 'db_close error: the db module is not available in the wasm build of blue'",
		}.String(),
	},
	{
		Name: "_db_exec",
		Fun: func(args ...Object) Object {
			return newError("`db_exec` error: the db module is not available in the wasm build of blue")
		},
		HelpStr: helpStrArgs{
			explanation: "`db_exec` is not available in the wasm build of blue",
			signature:   "db_exec(db: GoObj[*sql.DB], exec_query: str, exec_query_args: list[any]) -> Error",
			errors:      "CustomError",
			example:     "db_exec(db, 'CREATE TABLE ABC;', []) => 'db_exec error: the db module is not available in the wasm build of blue'",
		}.String(),
	},
	{
		Name: "_db_query",
		Fun: func(args ...Object) Object {
			return newError("`db_query` error: the db module is not available in the wasm build of blue")
		},
		HelpStr: helpStrArgs{
			explanation: "`db_query` is not available in the wasm build of blue",
			signature:   "db_query(db: GoObj[*sql.DB], query: str, query_args: list[any], named_cols: bool) -> Error",
			errors:      "CustomError",
			example:     "db_query(db, 'SELECT * FROM ABC;', [], false) => 'db_query error: the db module is not available in the wasm build of blue'",
		}.String(),
	},
}
