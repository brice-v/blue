//go:build !wasm

package object

import (
	"blue/consts"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/golang-module/carbon/v2"
	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"
)

var DbBuiltins = []*Builtin{
	{
		Name: "_db_open",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("db_open", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("db_open", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("db_open", 2, STRING_OBJ, args[1].Type())
			}
			dbName := args[0].(*Stringo).Value
			if dbName == "" {
				return newError("`db_open` error: db_name argument is empty")
			}
			dbType := args[1].(*Stringo).Value
			if dbType == "duckdb" && !duckdbAvailable {
				return newError("`db_open` error: %s. got=%s", consts.DuckDbLinkNote, dbType)
			}
			if dbType != "sqlite" && dbType != "duckdb" {
				return newError("`db_open` error: db_type must be 'sqlite' or 'duckdb'. got=%s", dbType)
			}
			db, err := sql.Open(dbType, dbName)
			if err != nil {
				return newError("`db_open` error: %s", err.Error())
			}
			return NewGoObj(db)
		},
		HelpStr: helpStrArgs{
			explanation: "`db_open` opens a connection to the builtin sqlite/duckdb db and returns the DB obj" + duckdbHelpNote,
			signature:   "db_open(db_name: str=':memory:', db_type: str='sqlite') -> GoObj[*sql.DB]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_open() => GoObj[*sql.DB]",
		}.String(),
	},
	{
		Name: "_db_ping",
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
	},
	{
		Name: "_db_close",
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
	},
	{
		Name: "_db_exec",
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
			if len(l) >= 1 {
				execArgs, err1 := getArgsAsGoObjs("db_exec", l)
				if err1 != nil {
					return newError("%s", err1.Error())
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
	},
	{
		Name: "_db_query",
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
				execArgs, err1 := getArgsAsGoObjs("db_query", l)
				if err1 != nil {
					return newError("%s", err1.Error())
				}
				rows, err = db.Value.Query(s, execArgs...)
			} else {
				rows, err = db.Value.Query(s)
			}
			if rows != nil {
				defer func() {
					err := rows.Close()
					if err != nil {
						log.Printf("Failed to close rows, error: %s", err.Error())
					}
				}()
			}
			if err != nil {
				return newError("`db_query` error: %s", err.Error())
			}
			colNames, err := rows.Columns()
			if err != nil {
				return newError("`db_query` error: %s", err.Error())
			}
			cols := make([]any, len(colNames))
			colPtrs := make([]any, len(colNames))
			for i := range colNames {
				colPtrs[i] = &cols[i]
			}
			returnList := &List{
				Elements: []Object{},
			}
			colTypes, err := rows.ColumnTypes()
			if err != nil {
				return newError("`db_query` error: %s", err.Error())
			}
			colDbTypes := make([]dbColumnKind, len(colTypes))
			for index := range colTypes {
				colDbTypes[index] = dbColumnTypeFor(colTypes, index)
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
					obj := createObjectFromDbInterface(e, colDbTypes[idx])
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
			if err := rows.Err(); err != nil {
				return newError("`db_query` error: %s", err.Error())
			}
			return returnList
		},
		HelpStr: helpStrArgs{
			explanation: "`db_query` is used to query the DB (ie. SELECT)",
			signature:   "db_query(db: GoObj[*sql.DB], query: str, query_args: list[any], named_cols: bool) -> list[any]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "db_query(db, 'SELECT * FROM ABC;', [], false) => list[any]",
		}.String(),
	},
}

func getArgsAsGoObjs(name string, l []Object) ([]any, error) {
	args := make([]any, len(l))
	for idx, e := range l {
		switch e.Type() {
		case STRING_OBJ:
			args[idx] = e.(*Stringo).Value
		case INTEGER_OBJ:
			args[idx] = e.(*Integer).Value
		case UINTEGER_OBJ:
			args[idx] = e.(*UInteger).Value
		case BIG_INTEGER_OBJ:
			args[idx] = e.(*BigInteger).Value
		case FLOAT_OBJ:
			args[idx] = e.(*Float).Value
		case BIG_FLOAT_OBJ:
			args[idx] = e.(*BigFloat).Value.String()
		case NULL_OBJ:
			args[idx] = nil
		case BOOLEAN_OBJ:
			args[idx] = e.(*Boolean).Value
		case BYTES_OBJ:
			args[idx] = e.(*Bytes).Value
		default:
			return nil, fmt.Errorf("argument list to `%s` included invalid type. got=%s", name, e.Type())
		}
	}
	return args, nil
}

type dbColumnType uint8

const (
	dbColUnknown dbColumnType = iota
	dbColInt
	dbColUInt
	dbColBigInt
	dbColFloat
	dbColString
	dbColUUID
	dbColBool
	dbColBytes
	dbColList
	dbColMap
	dbColDate
	dbColStruct
	dbColUnion
)

type dbColumnKind struct {
	base       dbColumnType
	name       string
	elem       *dbColumnKind
	fields     []string
	fieldKinds []*dbColumnKind
	key        *dbColumnKind
	val        *dbColumnKind
}

func dbColumnTypeFor(colTypes []*sql.ColumnType, index int) dbColumnKind {
	if index < 0 || index >= len(colTypes) {
		return dbColumnKind{base: dbColUnknown}
	}
	k := parseDbColumnTypeName(colTypes[index].DatabaseTypeName())
	if k.name == "" {
		k.name = strings.TrimSpace(colTypes[index].DatabaseTypeName())
	}
	return k
}

func parseDbColumnTypeName(typeName string) dbColumnKind {
	return parseDbColumnKind(typeName)
}

func parseDbColumnKind(typeName string) dbColumnKind {
	name := strings.TrimSpace(typeName)

	// Peel trailing list/array suffixes: INTEGER[], INTEGER[3], STRUCT(...)[], INTEGER[][] ...
	if strings.HasSuffix(name, "]") {
		open := strings.LastIndexByte(name, '[')
		if open != -1 && open < len(name)-1 {
			inner := strings.TrimSpace(name[:open])
			if inner != "" {
				elem := parseDbColumnKind(inner)
				return dbColumnKind{base: dbColList, elem: &elem}
			}
		}
	}

	upper := strings.ToUpper(name)
	if strings.HasPrefix(upper, "STRUCT") {
		open := strings.IndexByte(name, '(')
		if open == -1 {
			return dbColumnKind{base: dbColStruct}
		}
		closeIdx := matchingParen(name, open)
		if closeIdx == -1 {
			closeIdx = len(name) - 1
		}
		body := strings.TrimSpace(name[open+1 : closeIdx])
		fields := []string{}
		fieldKinds := []*dbColumnKind{}
		for _, part := range splitTopLevel(body, ',') {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			var fieldName string
			var typeStr string
			if strings.HasPrefix(part, "\"") {
				// quoted field name
				i := 1
				for i < len(part) {
					if part[i] == '"' {
						if i+1 < len(part) && part[i+1] == '"' {
							i += 2
							continue
						}
						break
					}
					i++
				}
				fieldName = part[1:i]
				fieldName = strings.ReplaceAll(fieldName, `""`, `"`)
				typeStr = strings.TrimSpace(part[i+1:])
			} else {
				toks := strings.Fields(part)
				if len(toks) >= 2 {
					fieldName = toks[0]
					typeStr = strings.TrimSpace(part[len(toks[0]):])
				} else {
					fieldName = ""
					typeStr = part
				}
			}
			if typeStr == "" {
				typeStr = "VARCHAR"
			}
			fk := parseDbColumnKind(typeStr)
			fields = append(fields, fieldName)
			fkCopy := fk
			fieldKinds = append(fieldKinds, &fkCopy)
		}
		return dbColumnKind{base: dbColStruct, fields: fields, fieldKinds: fieldKinds}
	}

	if strings.HasPrefix(upper, "MAP") {
		open := strings.IndexByte(name, '(')
		if open == -1 {
			return dbColumnKind{base: dbColMap}
		}
		closeIdx := matchingParen(name, open)
		if closeIdx == -1 {
			closeIdx = len(name) - 1
		}
		body := strings.TrimSpace(name[open+1 : closeIdx])
		parts := splitTopLevel(body, ',')
		var keyStr, valStr string
		if len(parts) >= 1 {
			keyStr = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			valStr = strings.TrimSpace(parts[1])
		}
		if keyStr == "" {
			keyStr = "VARCHAR"
		}
		if valStr == "" {
			valStr = "VARCHAR"
		}
		kk := parseDbColumnKind(keyStr)
		vv := parseDbColumnKind(valStr)
		return dbColumnKind{base: dbColMap, key: &kk, val: &vv}
	}

	if strings.HasPrefix(upper, "UNION") {
		open := strings.IndexByte(name, '(')
		fields := []string{}
		fieldKinds := []*dbColumnKind{}
		if open != -1 {
			closeIdx := matchingParen(name, open)
			if closeIdx == -1 {
				closeIdx = len(name) - 1
			}
			body := strings.TrimSpace(name[open+1 : closeIdx])
			for _, part := range splitTopLevel(body, ',') {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				var fieldName string
				var typeStr string
				if strings.HasPrefix(part, "\"") {
					i := 1
					for i < len(part) {
						if part[i] == '"' {
							if i+1 < len(part) && part[i+1] == '"' {
								i += 2
								continue
							}
							break
						}
						i++
					}
					fieldName = part[1:i]
					fieldName = strings.ReplaceAll(fieldName, `""`, `"`)
					typeStr = strings.TrimSpace(part[i+1:])
				} else {
					toks := strings.Fields(part)
					if len(toks) >= 2 {
						fieldName = toks[0]
						typeStr = strings.TrimSpace(part[len(toks[0]):])
					} else {
						fieldName = ""
						typeStr = part
					}
				}
				if typeStr == "" {
					typeStr = "VARCHAR"
				}
				fk := parseDbColumnKind(typeStr)
				fields = append(fields, fieldName)
				fkCopy := fk
				fieldKinds = append(fieldKinds, &fkCopy)
			}
		}
		return dbColumnKind{base: dbColUnion, fields: fields, fieldKinds: fieldKinds}
	}

	base := name
	if i := strings.IndexByte(base, '('); i != -1 {
		base = strings.TrimSpace(base[:i])
	}
	baseUpper := strings.ToUpper(base)
	switch baseUpper {
	case "TINYINT", "SMALLINT", "INT2", "INTEGER", "INT", "INT4", "SIGNED",
		"BIGINT", "INT8", "LONG", "MEDIUMINT", "SERIAL", "SMALLSERIAL", "BIGSERIAL":
		return dbColumnKind{base: dbColInt}
	case "UTINYINT", "USMALLINT", "UINTEGER", "UINT", "UINT8", "UINT16", "UINT32":
		return dbColumnKind{base: dbColUInt}
	case "HUGEINT", "UHUGEINT", "UBIGINT", "UINT64", "BIGNUM":
		return dbColumnKind{base: dbColBigInt}
	case "DECIMAL", "NUMERIC", "DEC":
		if decimalColumnType(name) == dbColBigInt {
			return dbColumnKind{base: dbColBigInt}
		}
		return dbColumnKind{base: dbColFloat}
	case "FLOAT", "REAL", "FLOAT4", "DOUBLE", "FLOAT8", "DOUBLE PRECISION":
		return dbColumnKind{base: dbColFloat}
	case "BOOLEAN", "BOOL":
		return dbColumnKind{base: dbColBool}
	case "BLOB", "BYTEA", "BINARY", "VARBINARY", "BYTES":
		return dbColumnKind{base: dbColBytes}
	case "UUID":
		return dbColumnKind{base: dbColUUID}
	case "VARCHAR", "TEXT", "STRING", "CHAR", "CHARACTER", "CHARACTER VARYING",
		"CLOB", "CITEXT", "BIT", "JSON", "JSONB", "ENUM",
		"TIME", "TIME_NS", "TIMETZ", "DATETIME", "TIMESTAMP",
		"TIMESTAMPTZ", "TIMESTAMP_S", "TIMESTAMP_MS", "TIMESTAMP_NS", "INTERVAL":
		return dbColumnKind{base: dbColString}
	case "DATE":
		return dbColumnKind{base: dbColDate}
	case "STRUCT":
		return dbColumnKind{base: dbColStruct}
	case "MAP":
		return dbColumnKind{base: dbColMap}
	case "LIST", "ARRAY":
		return dbColumnKind{base: dbColList}
	}
	return dbColumnKind{base: dbColUnknown}
}

func parseDecimalParams(name string) (width, scale int, ok bool) {
	open := strings.IndexByte(name, '(')
	end := strings.IndexByte(name, ')')
	if open == -1 || end <= open {
		return 0, 0, false
	}
	parts := strings.Split(name[open+1:end], ",")
	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, false
	}
	if len(parts) < 2 {
		return width, 0, true
	}
	scale, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, false
	}
	return width, scale, true
}

func decimalColumnType(name string) dbColumnType {
	width, scale, ok := parseDecimalParams(name)
	if ok && scale == 0 && width > 18 {
		return dbColBigInt
	}
	return dbColFloat
}

func matchingParen(s string, start int) int {
	depth := 0
	inQuote := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inQuote {
			if c == '"' {
				if i+1 < len(s) && s[i+1] == '"' {
					i++
				} else {
					inQuote = false
				}
			}
			continue
		}
		switch c {
		case '"':
			inQuote = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func splitTopLevel(s string, sep byte) []string {
	var parts []string
	depth := 0
	inQuote := false
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote {
			if c == '"' {
				if i+1 < len(s) && s[i+1] == '"' {
					i++
				} else {
					inQuote = false
				}
			}
			continue
		}
		switch c {
		case '"':
			inQuote = true
		case '(', '[':
			depth++
		case ')', ']':
			if depth > 0 {
				depth--
			}
		}
		if depth == 0 && c == sep {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	if start <= len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}

func createObjectFromDbInterface(input any, kind dbColumnKind) Object {
	if input == nil {
		return NULL
	}
	switch kind.base {
	case dbColString:
		if b, ok := input.([]byte); ok {
			return NewString(string(b))
		}
		if t, ok := input.(time.Time); ok {
			typeName := strings.ToUpper(strings.TrimSpace(kind.name))
			switch typeName {
			case "TIME":
				return NewString(carbon.CreateFromStdTime(t).ToTimeString())
			case "TIME_NS":
				return NewString(carbon.CreateFromStdTime(t).ToTimeNanoString())
			case "TIMETZ":
				return NewString(t.Format("15:04:05-07"))
			case "TIMESTAMP":
				return NewString(carbon.CreateFromStdTime(t).ToDateTimeMicroString())
			case "TIMESTAMPTZ":
				return NewString(carbon.CreateFromStdTime(t).ToDateTimeString())
			case "TIMESTAMP_NS":
				return NewString(carbon.CreateFromStdTime(t).ToDateTimeNanoString())
			}
		}
	case dbColUUID:
		switch x := input.(type) {
		case []byte:
			if len(x) != 16 {
				return NewString(string(x))
			}
			var id duckdb.UUID
			copy(id[:], x)
			return NewString(id.String())
		case string:
			return NewString(x)
		}
	case dbColBigInt:
		switch x := input.(type) {
		case uint64:
			return &BigInteger{Value: new(big.Int).SetUint64(x)}
		case uint32:
			return &BigInteger{Value: new(big.Int).SetUint64(uint64(x))}
		case duckdb.Decimal:
			return &BigInteger{Value: new(big.Int).Set(x.Value)}
		case string:
			if bi, ok := new(big.Int).SetString(x, 10); ok {
				return &BigInteger{Value: bi}
			}
		}
	case dbColFloat:
		if x, ok := input.(duckdb.Decimal); ok {
			return &Float{Value: x.Float64()}
		}
	case dbColBool:
		if x, ok := input.(int64); ok {
			return &Boolean{Value: x != 0}
		}
	case dbColDate:
		switch x := input.(type) {
		case time.Time:
			return NewString(carbon.CreateFromStdTime(x).ToDateString())
		case []byte:
			return NewString(string(x))
		case string:
			return NewString(x)
		}
	case dbColStruct:
		if x, ok := input.(map[string]any); ok {
			fields := kind.fields
			fieldKinds := kind.fieldKinds
			if len(fields) == 0 {
				keys := make([]string, 0, len(x))
				for k := range x {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				fields = keys
				fieldKinds = make([]*dbColumnKind, len(keys))
				for i := range fieldKinds {
					kk := dbColumnKind{base: dbColUnknown}
					fieldKinds[i] = &kk
				}
			}
			// keep field order stable so struct equality is order independent
			if len(fields) > 1 {
				order := make([]int, len(fields))
				for i := range order {
					order[i] = i
				}
				sort.Slice(order, func(a, b int) bool {
					return fields[order[a]] < fields[order[b]]
				})
				sortedFields := make([]string, len(fields))
				sortedKinds := make([]*dbColumnKind, len(fields))
				for i, idx := range order {
					sortedFields[i] = fields[idx]
					if idx < len(fieldKinds) {
						sortedKinds[i] = fieldKinds[idx]
					}
				}
				fields = sortedFields
				fieldKinds = sortedKinds
			}
			values := make([]Object, len(fields))
			for i, f := range fields {
				kk := dbColumnKind{base: dbColUnknown}
				if i < len(fieldKinds) && fieldKinds[i] != nil {
					kk = *fieldKinds[i]
				}
				v, ok := x[f]
				if !ok || v == nil {
					values[i] = NULL
				} else {
					values[i] = createObjectFromDbInterface(v, kk)
				}
			}
			obj, err := NewBlueStruct(fields, values)
			if err != nil {
				return newError("struct conversion error: %s", err.Error())
			}
			return obj
		}
	case dbColList:
		if x, ok := input.([]any); ok {
			objs := &List{Elements: make([]Object, len(x))}
			ek := dbColumnKind{base: dbColUnknown}
			if kind.elem != nil {
				ek = *kind.elem
			}
			for i, e := range x {
				objs.Elements[i] = createObjectFromDbInterface(e, ek)
			}
			return objs
		}
	case dbColMap:
		if x, ok := input.(duckdb.OrderedMap); ok {
			m := &Map{Pairs: NewPairsMapWithSize(x.Len())}
			kk := dbColumnKind{base: dbColUnknown}
			if kind.key != nil {
				kk = *kind.key
			}
			vv := dbColumnKind{base: dbColUnknown}
			if kind.val != nil {
				vv = *kind.val
			}
			for _, k := range x.Keys() {
				kAsObj := createObjectFromDbInterface(k, kk)
				v, _ := x.Get(k)
				vAsObj := createObjectFromDbInterface(v, vv)
				hk := HashKey{Type: kAsObj.Type(), Value: HashObject(kAsObj)}
				m.Pairs.Set(hk, MapPair{Key: kAsObj, Value: vAsObj})
			}
			return m
		}
		if x, ok := input.(map[string]any); ok {
			m := &Map{Pairs: NewPairsMapWithSize(len(x))}
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			vv := dbColumnKind{base: dbColUnknown}
			if kind.val != nil {
				vv = *kind.val
			}
			for _, k := range keys {
				v := x[k]
				vvAsObj := createObjectFromDbInterface(v, vv)
				kk := &Stringo{Value: k}
				hk := HashKey{Type: STRING_OBJ, Value: HashObject(kk)}
				m.Pairs.Set(hk, MapPair{Key: kk, Value: vvAsObj})
			}
			return m
		}
	case dbColUnion:
		switch x := input.(type) {
		case duckdb.Union:
			return createObjectFromDbInterface(x.Value, dbColumnKind{base: dbColUnknown})
		}
	}
	switch x := input.(type) {
	case int64:
		return NewInteger(x)
	case int32:
		return NewInteger(int64(x))
	case int16:
		return NewInteger(int64(x))
	case int8:
		return NewInteger(int64(x))
	case uint64:
		return &UInteger{Value: x}
	case uint32:
		return &UInteger{Value: uint64(x)}
	case uint16:
		return &UInteger{Value: uint64(x)}
	case uint8:
		return &UInteger{Value: uint64(x)}
	case string:
		return NewString(x)
	case float64:
		return &Float{Value: x}
	case float32:
		s := strconv.FormatFloat(float64(x), 'g', -1, 32)
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return newError("float conversion error: %s", err.Error())
		}
		return &Float{Value: f}
	case []byte:
		return &Bytes{Value: x}
	case bool:
		return &Boolean{Value: x}
	case time.Time:
		// TODO: Return time.Time as GoObj if this starts to be better supported
		return &Stringo{Value: carbon.CreateFromStdTime(x).ToDateTimeMilliString()}
	case duckdb.Decimal:
		// TODO: This is potentially inefficient
		bf, err := decimal.NewFromString(x.String())
		if err != nil {
			return newError("decimal conversion error: %s", err.Error())
		}
		return &BigFloat{Value: bf}
	case *big.Int:
		return &BigInteger{Value: x}
	case []any:
		objs := &List{Elements: make([]Object, len(x))}
		for i, e := range x {
			objs.Elements[i] = createObjectFromDbInterface(e, dbColumnKind{base: dbColUnknown})
		}
		return objs
	case duckdb.OrderedMap:
		m := &Map{Pairs: NewPairsMapWithSize(x.Len())}
		for _, k := range x.Keys() {
			kAsObj := createObjectFromDbInterface(k, dbColumnKind{base: dbColUnknown})
			v, _ := x.Get(k)
			vAsObj := createObjectFromDbInterface(v, dbColumnKind{base: dbColUnknown})
			hk := HashKey{Type: kAsObj.Type(), Value: HashObject(kAsObj)}
			m.Pairs.Set(hk, MapPair{Key: kAsObj, Value: vAsObj})
		}
		return m
	case map[string]any:
		m := &Map{Pairs: NewPairsMapWithSize(len(x))}
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := x[k]
			vv := createObjectFromDbInterface(v, dbColumnKind{base: dbColUnknown})
			kk := &Stringo{Value: k}
			hk := HashKey{Type: STRING_OBJ, Value: HashObject(kk)}
			m.Pairs.Set(hk, MapPair{Key: kk, Value: vv})
		}
		return m
	case duckdb.Union:
		return createObjectFromDbInterface(x.Value, dbColumnKind{base: dbColUnknown})
	case duckdb.Interval:
		m := NewOrderedMapWithSize[string, Object](3)
		m.Set("days", NewInteger(int64(x.Days)))
		m.Set("micros", NewInteger(x.Micros))
		m.Set("months", NewInteger(int64(x.Months)))
		m.Set("t", NewString("duckdb.Interval"))
		return CreateMapObjectForGoMap(*m)
	default:
		return nil
	}
}
