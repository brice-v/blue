package evaluator

import (
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	_ "modernc.org/sqlite"
)

// StdModFileAndBuiltins keeps the file and builtins together for each std lib module
type StdModFileAndBuiltins struct {
	File     string         // File is the actual code used for the module
	Builtins BuiltinMapType // Builtins is the map of functions to be used by the module
}

//go:embed std/http.b
var stdHttpFile string

//go:embed std/time.b
var stdTimeFile string

//go:embed std/search.b
var stdSearchFile string

//go:embed std/db.b
var stdDbFile string

// TODO: Could use an embed.FS and read the files that way rather then set each one individually
// but it works well enough for now (if we used embed.FS we probably just need a helper)
var _std_mods = map[string]StdModFileAndBuiltins{
	"http":   {File: stdHttpFile, Builtins: _http_builtin_map},
	"time":   {File: stdTimeFile, Builtins: _time_builtin_map},
	"search": {File: stdSearchFile, Builtins: _search_builtin_map},
	"db":     {File: stdDbFile, Builtins: _db_builtin_map},
}

func (e *Evaluator) IsStd(name string) bool {
	_, ok := _std_mods[name]
	return ok
}

func (e *Evaluator) AddStdLibToEnv(name string) {
	if !e.IsStd(name) {
		fmt.Printf("AddStdLibToEnv: '%s' is not in std lib map\n", name)
		os.Exit(1)
	}
	fb := _std_mods[name]
	l := lexer.New(fb.File)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("ParserError in `%s` module: %s\n", name, msg)
		}
		os.Exit(1)
	}
	newE := New()
	newE.Builtins.PushBack(fb.Builtins)
	val := newE.Eval(program)
	if isError(val) {
		fmt.Printf("EvaluatorError in `%s` module: %s\n", name, val.(*object.Error).Message)
		os.Exit(1)
	}
	mod := &object.Module{Name: name, Env: newE.env}
	e.env.Set(name, mod)
}

// Note: Look at how we import the get function in http.b
var _http_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_get": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`get` expects 1 argument. got %d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `get` must be STRING. got %s", args[0].Type())
			}
			resp, err := http.Get(args[0].(*object.Stringo).Value)
			if err != nil {
				return newError("`get` failed: %s", err.Error())
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("`get` failed: %s", err.Error())
			}
			return &object.Stringo{Value: string(body)}
		},
	},
})

var _time_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_sleep": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`sleep` expects 1 argument. got %d", len(args))
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("argument to `sleep` must be INTEGER, got %s", args[0].Type())
			}
			i := args[0].(*object.Integer).Value
			if i < 0 {
				return newError("INTEGER argument to `sleep` must be > 0, got %d", i)
			}
			time.Sleep(time.Duration(i) * time.Millisecond)
			return NULL
		},
	},
	"_now": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("`now` expects 0 arguments. got %d", len(args))
			}
			return &object.Integer{Value: time.Now().Unix()}
		},
	},
})

var _search_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_by_xpath": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`by_xpath` expects 3 arguments. got %d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `by_xpath` should be STRING. got %s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `by_xpath` should be STRING. got %s", args[1].Type())
			}
			if args[2].Type() != object.BOOLEAN_OBJ {
				return newError("argument 3 to `by_xpath` should be BOOLEAN. got %s", args[2].Type())
			}
			strToSearch := args[0].(*object.Stringo).Value
			if strToSearch == "" {
				return newError("`by_xpath` error: str_to_search argument is empty")
			}
			strQuery := args[1].(*object.Stringo).Value
			if strQuery == "" {
				return newError("`by_xpath` error: query argument is empty")
			}
			shouldFindOne := args[2].(*object.Boolean).Value
			doc, err := htmlquery.Parse(strings.NewReader(strToSearch))
			if err != nil {
				return newError("`by_xpath` failed to parse document as html: error %s", err.Error())
			}
			if !shouldFindOne {
				listToReturn := &object.List{Elements: make([]object.Object, 0)}
				for _, e := range htmlquery.Find(doc, strQuery) {
					result := htmlquery.OutputHTML(e, true)
					listToReturn.Elements = append(listToReturn.Elements, &object.Stringo{Value: result})
				}
				return listToReturn
			} else {
				e := htmlquery.FindOne(doc, strQuery)
				result := htmlquery.OutputHTML(e, true)
				return &object.Stringo{Value: result}
			}
		},
	},
	"_by_regex": {
		Fun: func(args ...object.Object) object.Object {
			return NULL
		},
	},
})

var _db_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_open": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`open` expects 1 argument. got %d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `open` should be STRING. got %s", args[0].Type())
			}
			dbName := args[0].(*object.Stringo).Value
			if dbName == "" {
				return newError("`open` error: db_name argument is empty")
			}
			db, err := sql.Open("sqlite", dbName)
			if err != nil {
				return newError("`open` error: %s", err.Error())
			}
			curDB := dbCount.Add(1)
			DBMap.Put(curDB, db)
			return object.CreateBasicMapObject("DB", curDB)
		},
	},
	"_ping": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`ping` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("argument 1 to `ping` should be INTEGER. got=%s", args[0].Type())
			}
			i := args[0].(*object.Integer).Value
			if db, ok := DBMap.Get(i); ok {
				err := db.Ping()
				if err != nil {
					return &object.Stringo{Value: err.Error()}
				}
				return NULL
			}
			return newError("`ping` error: DB not found")
		},
	},
	"_close": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`close` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("argument 1 to `close` should be INTEGER. got=%s", args[0].Type())
			}
			i := args[0].(*object.Integer).Value
			if db, ok := DBMap.Get(i); ok {
				err := db.Close()
				if err != nil {
					return newError("`close` error: %s", err.Error())
				}
				DBMap.Remove(i)
				dbCount.Store(dbCount.Load() - 1)
				return NULL
			}
			return newError("`close` error: DB not found")
		},
	},
	"_exec": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`exec` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("argument 1 to `exec` should be INTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `exec` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.LIST_OBJ {
				return newError("argument 3 to `exec` should be LIST. got=%s", args[2].Type())
			}
			i := args[0].(*object.Integer).Value
			s := args[1].(*object.Stringo).Value
			l := args[2].(*object.List).Elements
			if db, ok := DBMap.Get(i); ok {
				var result sql.Result
				var err error
				if len(l) > 1 {
					execArgs := make([]any, len(l))
					for idx, e := range l {
						// TODO: Type checking l elements for exec? (allow not only string)
						if e.Type() != object.STRING_OBJ {
							return newError("argument list to `exec` should all be STRING. got=%s", e.Type())
						}
						execArgs[idx] = e.(*object.Stringo).Value
					}
					result, err = db.Exec(s, execArgs...)
				} else {
					result, err = db.Exec(s)
				}
				if err != nil {
					return newError("`exec` error: %s", err.Error())
				}
				lastInsertId, err := result.LastInsertId()
				if err != nil {
					return newError("`exec` error: %s", err.Error())
				}
				rowsAffected, err := result.RowsAffected()
				if err != nil {
					return newError("`exec` error: %s", err.Error())
				}
				mapToConvert := object.NewOrderedMap[string, object.Object]()
				mapToConvert.Set("last_insert_id", &object.Integer{Value: lastInsertId})
				mapToConvert.Set("rows_affected", &object.Integer{Value: rowsAffected})
				return object.CreateMapObjectForGoMap(*mapToConvert)
			}
			return newError("`exec` error: DB not found")
		},
	},
	"_query": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newError("`query` expects 4 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("argument 1 to `query` should be INTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `query` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.LIST_OBJ {
				return newError("argument 3 to `query` should be LIST. got=%s", args[2].Type())
			}
			if args[3].Type() != object.BOOLEAN_OBJ {
				return newError("argument 4 to `query` should be BOOLEAN. got=%s", args[3].Type())
			}
			i := args[0].(*object.Integer).Value
			s := args[1].(*object.Stringo).Value
			l := args[2].(*object.List).Elements
			isNamedCols := args[3].(*object.Boolean).Value
			if db, ok := DBMap.Get(i); ok {
				var rows *sql.Rows
				var err error
				if len(l) > 1 {
					execArgs := make([]any, len(l))
					for idx, e := range l {
						// TODO: Type checking l elements for query? (allow not only string)
						if e.Type() != object.STRING_OBJ {
							return newError("argument list to `query` should all be STRING. got=%s", e.Type())
						}
						execArgs[idx] = e.(*object.Stringo).Value
					}
					rows, err = db.Query(s, execArgs...)
				} else {
					rows, err = db.Query(s)
				}
				defer rows.Close()
				if err != nil {
					return newError("`query` error: %s", err.Error())
				}
				colNames, err := rows.Columns()
				if err != nil {
					return newError("`query` error: %s", err.Error())
				}
				// Get Types to properly scan
				// https://www.sqlite.org/datatype3.html
				// NULL. The value is a NULL value.
				// INTEGER. The value is a signed integer, stored in 0, 1, 2, 3, 4, 6, or 8 bytes depending on the magnitude of the value.
				// REAL. The value is a floating point value, stored as an 8-byte IEEE floating point number.
				// TEXT. The value is a text string, stored using the database encoding (UTF-8, UTF-16BE or UTF-16LE).
				// BLOB. The value is a blob of data, stored exactly as it was input.
				cols := make([]interface{}, len(colNames))
				colPtrs := make([]interface{}, len(colNames))
				for i := 0; i < len(colNames); i++ {
					colPtrs[i] = &cols[i]
				}
				returnList := &object.List{
					Elements: make([]object.Object, 0),
				}
				for rows.Next() {
					err = rows.Scan(colPtrs...)
					if err != nil {
						return newError("`query` error: %s", err.Error())
					}
					rowList := &object.List{
						Elements: make([]object.Object, len(cols)),
					}
					rowMap := object.NewOrderedMap[string, object.Object]()
					for idx, e := range cols {
						obj := object.CreateObjectFromDbInterface(e)
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
						returnList.Elements = append(returnList.Elements, object.CreateMapObjectForGoMap(*rowMap))
					}
				}
				return returnList
			}
			return newError("`exec` error: DB not found")
		},
	},
})
