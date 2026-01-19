package object

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var DbBuiltins = NewBuiltinSliceType{
	{Name: "_db_open", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("db_open", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("db_open", 1, STRING_OBJ, args[0].Type())
			}
			dbName := args[0].(*Stringo).Value
			if dbName == "" {
				return newError("`db_open` error: db_name argument is empty")
			}
			db, err := sql.Open("sqlite", dbName)
			if err != nil {
				return newError("`db_open` error: %s", err.Error())
			}
			return NewGoObj(db)
		},
		HelpStr: helpStrArgs{
			explanation: "`db_open` opens a connection to the builtin sqlite db and returns the DB obj",
			signature:   "db_open(db_name: str=':memory:') -> GoObj[*sql.DB]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_open() => GoObj[*sql.DB]",
		}.String(),
	}},
	{Name: "_db_ping", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("db_ping", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("db_ping", 1, GO_OBJ, args[0].Type())
			}
			db, ok := args[0].(*GoObj[*sql.DB])
			if !ok {
				return newPositionalTypeErrorForGoObj("db_ping", 1, "*sql.DB", args[0])
			}
			err := db.Value.Ping()
			if err != nil {
				return &Stringo{Value: err.Error()}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`db_ping` pings the connection to the DB to verify connectivity. if no error, null is returned",
			signature:   "db_ping(db: GoObj[*sql.DB]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_ping(db) => null",
		}.String(),
	}},
	{Name: "_db_close", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("db_close", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("db_close", 1, GO_OBJ, args[0].Type())
			}
			db, ok := args[0].(*GoObj[*sql.DB])
			if !ok {
				return newPositionalTypeErrorForGoObj("db_close", 1, "*sql.DB", args[0])
			}
			err := db.Value.Close()
			if err != nil {
				return newError("`db_close` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`db_close` closes the connection to the DB",
			signature:   "db_close(db: GoObj[*sql.DB]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_close(db) => null",
		}.String(),
	}},
	{Name: "_db_exec", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("db_exec", len(args), 3, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("db_exec", 1, GO_OBJ, args[0].Type())
			}
			db, ok := args[0].(*GoObj[*sql.DB])
			if !ok {
				return newPositionalTypeErrorForGoObj("db_exec", 1, "*sql.DB", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("db_exec", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != LIST_OBJ {
				return newPositionalTypeError("db_exec", 3, LIST_OBJ, args[2].Type())
			}
			s := args[1].(*Stringo).Value
			l := args[2].(*List).Elements

			var result sql.Result
			var err error
			if len(l) > 1 {
				execArgs := make([]any, len(l))
				for idx, e := range l {
					switch e.Type() {
					case STRING_OBJ:
						execArgs[idx] = e.(*Stringo).Value
					case INTEGER_OBJ:
						execArgs[idx] = e.(*Integer).Value
					case FLOAT_OBJ:
						execArgs[idx] = e.(*Float).Value
					case NULL_OBJ:
						execArgs[idx] = nil
					case BOOLEAN_OBJ:
						execArgs[idx] = e.(*Boolean).Value
					case BYTES_OBJ:
						execArgs[idx] = e.(*Bytes).Value
					default:
						return newError("argument list to `db_exec` included invalid type. got=%s", e.Type())
					}
				}
				result, err = db.Value.Exec(s, execArgs...)
			} else {
				result, err = db.Value.Exec(s)
			}
			if err != nil {
				return newError("`db_exec` error: %s", err.Error())
			}
			lastInsertId, err := result.LastInsertId()
			if err != nil {
				return newError("`db_exec` error: %s", err.Error())
			}
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return newError("`db_exec` error: %s", err.Error())
			}
			mapToConvert := NewOrderedMap[string, Object]()
			mapToConvert.Set("last_insert_id", &Integer{Value: lastInsertId})
			mapToConvert.Set("rows_affected", &Integer{Value: rowsAffected})
			return CreateMapObjectForGoMap(*mapToConvert)
		},
		HelpStr: helpStrArgs{
			explanation: "`db_exec` is used to execute queries against the DB that affect rows (ie. INSERT statments)",
			signature:   "db_exec(db: GoObj[*sql.DB], exec_query: str, exec_query_args: list[any]) -> {last_insert_id: _, rows_affected: _}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_exec(db, 'CREATE TABLE ABC;', []) => {last_insert_id: 1, rows_affected: 1}",
		}.String(),
	}},
	{Name: "_db_query", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("db_query", len(args), 4, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("db_query", 1, GO_OBJ, args[0].Type())
			}
			db, ok := args[0].(*GoObj[*sql.DB])
			if !ok {
				return newPositionalTypeErrorForGoObj("db_query", 1, "*sql.DB", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("db_query", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != LIST_OBJ {
				return newPositionalTypeError("db_query", 3, LIST_OBJ, args[2].Type())
			}
			if args[3].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("db_query", 4, BOOLEAN_OBJ, args[3].Type())
			}
			s := args[1].(*Stringo).Value
			l := args[2].(*List).Elements
			isNamedCols := args[3].(*Boolean).Value
			var rows *sql.Rows
			var err error
			if len(l) >= 1 {
				execArgs := make([]any, len(l))
				for idx, e := range l {
					switch e.Type() {
					case STRING_OBJ:
						execArgs[idx] = e.(*Stringo).Value
					case INTEGER_OBJ:
						execArgs[idx] = e.(*Integer).Value
					case FLOAT_OBJ:
						execArgs[idx] = e.(*Float).Value
					case NULL_OBJ:
						execArgs[idx] = nil
					case BOOLEAN_OBJ:
						execArgs[idx] = e.(*Boolean).Value
					case BYTES_OBJ:
						execArgs[idx] = e.(*Bytes).Value
					default:
						return newError("argument list to `db_query` included invalid type. got=%s", e.Type())
					}
				}
				rows, err = db.Value.Query(s, execArgs...)
			} else {
				rows, err = db.Value.Query(s)
			}
			if rows != nil {
				defer rows.Close()
			}
			if err != nil {
				return newError("`db_query` error: %s", err.Error())
			}
			colNames, err := rows.Columns()
			if err != nil {
				return newError("`db_query` error: %s", err.Error())
			}
			// Get Types to properly scan
			// https://www.sqlite.org/datatype3.html
			// NULL. The value is a NULL value.
			// INTEGER. The value is a signed integer, stored in 0, 1, 2, 3, 4, 6, or 8 bytes depending on the magnitude of the value.
			// REAL. The value is a floating point value, stored as an 8-byte IEEE floating point number.
			// TEXT. The value is a text string, stored using the database encoding (UTF-8, UTF-16BE or UTF-16LE).
			// BLOB. The value is a blob of data, stored exactly as it was input.
			cols := make([]any, len(colNames))
			colPtrs := make([]any, len(colNames))
			for i := range colNames {
				colPtrs[i] = &cols[i]
			}
			returnList := &List{
				Elements: []Object{},
			}
			for rows.Next() {
				err = rows.Scan(colPtrs...)
				if err != nil {
					return newError("`db_query` error: %s", err.Error())
				}
				rowList := &List{
					Elements: make([]Object, len(cols)),
				}
				var rowMap *OrderedMap2[string, Object] = nil
				if isNamedCols {
					rowMap = NewOrderedMap[string, Object]()
				}
				for idx, e := range cols {
					obj := CreateObjectFromDbInterface(e)
					if obj == nil {
						obj = NULL
					}
					if !isNamedCols {
						rowList.Elements[idx] = obj
					} else {
						rowMap.Set(colNames[idx], obj)
					}
				}
				if !isNamedCols {
					returnList.Elements = append(returnList.Elements, rowList)
				} else {
					returnList.Elements = append(returnList.Elements, CreateMapObjectForGoMap(*rowMap))
				}
			}
			return returnList
		},
		HelpStr: helpStrArgs{
			explanation: "`db_query` is used to query the DB (ie. SELECT)",
			signature:   "db_query(db: GoObj[*sql.DB], query: str, query_args: list[any], named_cols: bool) -> list[any]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_query(db, 'SELECT * FROM ABC;', [], false) => list[any]",
		}.String(),
	}},
}
