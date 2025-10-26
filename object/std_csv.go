package object

import (
	"encoding/csv"
	"strings"
	"unicode/utf8"
)

var CsvBuiltins = NewBuiltinSliceType{
	{Name: "_parse", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 6 {
				return newInvalidArgCountError("parse", len(args), 6, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("parse", 1, STRING_OBJ, args[0].Type())
			}
			// parse(data, delimeter=',', named_fields=false, comment=null, lazy_quotes=false, trim_leading_space=false) {
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("parse", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("parse", 3, BOOLEAN_OBJ, args[2].Type())
			}
			if args[3].Type() != NULL_OBJ && args[3].Type() != STRING_OBJ {
				return newPositionalTypeError("parse", 4, "NULL or STRING", args[3].Type())
			}
			if args[4].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("parse", 5, BOOLEAN_OBJ, args[4].Type())
			}
			if args[5].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("parse", 6, BOOLEAN_OBJ, args[5].Type())
			}
			data := args[0].(*Stringo).Value
			delimeter := args[1].(*Stringo).Value
			namedFields := args[2].(*Boolean).Value
			useComment := false
			var comment rune
			if args[3].Type() == NULL_OBJ {
				useComment = true
			} else {
				c := args[3].(*Stringo).Value
				if utf8.RuneCountInString(c) != 1 {
					return newError("parse error: comment length is not 1. got=%d '%s'", utf8.RuneCountInString(c), c)
				}
				comment = []rune(c)[0]
			}
			lazyQuotes := args[4].(*Boolean).Value
			trimLeadingSpace := args[5].(*Boolean).Value
			if utf8.RuneCountInString(delimeter) != 1 {
				return newError("parse error: delimeter length is not 1. got=%d '%s'", utf8.RuneCountInString(delimeter), delimeter)
			}
			dRune := []rune(delimeter)[0]

			reader := csv.NewReader(strings.NewReader(data))
			reader.Comma = dRune
			if useComment {
				reader.Comment = comment
			}
			reader.LazyQuotes = lazyQuotes
			reader.TrimLeadingSpace = trimLeadingSpace

			rows, err := reader.ReadAll()
			if err != nil {
				return newError("parse error: %s", err.Error())
			}
			if !namedFields {
				// Here we are just returning a list of lists
				allRows := &List{
					Elements: make([]Object, len(rows)),
				}
				for i, row := range rows {
					rowList := &List{
						Elements: make([]Object, len(row)),
					}
					for j, e := range row {
						rowList.Elements[j] = &Stringo{Value: e}
					}
					allRows.Elements[i] = rowList
				}
				return allRows
			}

			if len(rows) < 1 {
				return newError("parse error: named fields requires at least 1 row in the csv to act as the header")
			}
			headerRow := rows[0]
			rows = rows[1:]
			allRows := &List{
				Elements: make([]Object, len(rows)),
			}
			for i, row := range rows {
				if len(row) != len(headerRow) {
					return newError("parse error: row length did not match header row length. got=%d, want=%d", len(row), len(headerRow))
				}
				m := NewOrderedMap[string, Object]()
				for i, v := range row {
					m.Set(headerRow[i], &Stringo{Value: v})
				}
				allRows.Elements[i] = CreateMapObjectForGoMap(*m)
			}
			return allRows
		},
		HelpStr: helpStrArgs{
			explanation: "`parse` parses the string or bytes as a CSV and returns the data as a list of objects",
			signature:   "parse(data: str|bytes, delimeter: str=',', named_fields: bool=false, comment: str|null=null, lazy_quotes: bool=false, trim_leading_space: bool=false) -> list[any]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "parse(data) => list[any]",
		}.String(),
	}},
	{Name: "_dump", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("dump", len(args), 3, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("dump", 1, LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("dump", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("dump", 3, BOOLEAN_OBJ, args[2].Type())
			}
			l := args[0].(*List).Elements
			comma := args[1].(*Stringo).Value
			if utf8.RuneCountInString(comma) != 1 {
				return newError("dump error: comma needs to be 1 character long. got=%d", utf8.RuneCountInString(comma))
			}
			c := []rune(comma)[0]
			useCrlf := args[2].(*Boolean).Value
			if len(l) < 1 {
				return newError("dump error: list was empty. got=%d", len(l))
			}
			if l[0].Type() != MAP_OBJ && l[0].Type() != LIST_OBJ {
				return newError("dump error: list should be a list of maps, or list of lists. got=%s", l[0].Type())
			}
			offset := 0
			if l[0].Type() == MAP_OBJ {
				// Account for headers
				offset = 1
			}
			allRows := make([][]string, len(l)+offset)

			// checking types and info
			if l[0].Type() == MAP_OBJ {
				var keys []HashKey
				for i, e := range l {
					if e.Type() != MAP_OBJ {
						return newError("dump error: invalid data. for rows that should be MAPs, found %s", e.Type())
					}
					// Validate that all the keys are at least the same - then we can use inspect
					// to get the actual keys and also use inspect for all the values
					// May just want to use a separate loops
					mps := e.(*Map).Pairs
					if keys == nil && i == 0 {
						keys = append(keys, mps.Keys...)
						for _, k := range mps.Keys {
							mp, _ := mps.Get(k)
							// This is for the headers
							allRows[i] = append(allRows[i], mp.Key.Inspect())
							allRows[i+offset] = append(allRows[i+offset], mp.Value.Inspect())
						}
					} else {
						if len(keys) != len(mps.Keys) {
							return newError("dump error: invalid data. found a row where number of keys did not match")
						}
						for j, k := range mps.Keys {
							if keys[j] != k {
								return newError("dump error: invalid data. found a row where the key at a certain position did not match the expected")
							}
							mp, _ := mps.Get(k)
							allRows[i+offset] = append(allRows[i+offset], mp.Value.Inspect())
						}
					}
				}
			} else {
				for i, e := range l {
					if e.Type() != LIST_OBJ {
						return newError("dump error: invalid data. for rows that should be LISTs, found %s", e.Type())
					}
					rowL := e.(*List).Elements
					for _, elem := range rowL {
						// No offset should be needed here (but if we added it, it would just be 0)
						allRows[i] = append(allRows[i], elem.Inspect())
					}
				}
			}
			sb := &strings.Builder{}
			w := csv.NewWriter(sb)
			w.Comma = c
			w.UseCRLF = useCrlf
			err := w.WriteAll(allRows)
			if err != nil {
				return newError("dump error: csv writer error: %s", err.Error())
			}
			return &Stringo{Value: sb.String()}
		},
		HelpStr: helpStrArgs{
			explanation: "`dump` dumps the data to a CSV",
			signature:   "dump(data: list[any], comma: str=',', use_crlf: bool=false) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dump(data) => null",
		}.String(),
	}},
}
